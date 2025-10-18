package main

import (
	"bufio"
	"flag"
	"fmt"
	"go.bug.st/serial"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	DefaultDelay    = 20
	DelayMin        = 10
	DelayMax        = 30
	SafetyPositions = "90 45 180 180 90 10"
)

var validSerialSpeeds = []int{9600, 19200, 38400, 57600, 115200}

func isValidSerialSpeed(speed int) bool {
	for _, s := range validSerialSpeeds {
		if s == speed {
			return true
		}
	}
	return false
}

type Motor struct {
	Name     string
	Min      int
	Max      int
	Position int
}

type RobotState struct {
	Base         Motor
	Shoulder     Motor
	Elbow        Motor
	Wrist        Motor
	WristRotate  Motor
	Gripper      Motor
	Delay        int
	Sim          bool
	Connected    bool
	SerialPort   string
	SerialSpeed  int
	SerialHandle serial.Port
}

func NewRobotState(sim bool) *RobotState {
	return &RobotState{
		Base:        Motor{"Base", 0, 180, 90},
		Shoulder:    Motor{"Shoulder", 15, 165, 45},
		Elbow:       Motor{"Elbow", 0, 180, 180},
		Wrist:       Motor{"Wrist", 0, 180, 180},
		WristRotate: Motor{"WristRotate", 0, 180, 90},
		Gripper:     Motor{"Gripper", 10, 73, 10},
		Delay:       DefaultDelay,
		Sim:         sim,
		Connected:   false,
	}
}

func (r *RobotState) Status() string {
	if !r.Connected {
		return "STATUS DISCONNECTED"
	}
	return fmt.Sprintf("STATUS M1:%d M2:%d M3:%d M4:%d M5:%d M6:%d DELAY:%d",
		r.Base.Position, r.Shoulder.Position, r.Elbow.Position, r.Wrist.Position, r.WristRotate.Position, r.Gripper.Position, r.Delay)
}

func main() {
	port := flag.Int("port", 8999, "TCP port to listen on")
	simMode := flag.Bool("sim", false, "Enable simulation mode")
	flag.Parse()

	if *simMode {
		log.Println("Simulation mode enabled.")
	}
	state := NewRobotState(*simMode)

	addr := fmt.Sprintf(":%d", *port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Server listening on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConnection(conn, state)
	}
}

func handleConnection(conn net.Conn, state *RobotState) {
	defer conn.Close()
	log.Printf("Connection from %s", conn.RemoteAddr())
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		resp := handleCommand(line, state)
		log.Printf("Command: %s | Response: %s", line, resp)
		fmt.Fprintln(conn, resp)
		if resp == "BYE" {
			break
		}
	}
}

func handleCommand(cmd string, state *RobotState) string {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return "INVALID COMMAND FORMAT"
	}
	switch fields[0] {
	case "CONNECT":
		if state.Connected {
			if state.SerialHandle != nil {
				state.SerialHandle.Close()
				state.SerialHandle = nil
			}
			state.Connected = false
		}
		if state.Sim {
			state.Connected = true
			return "CONNECTED"
		}
		if len(fields) < 3 {
			return "ERROR Missing port or speed"
		}
		port := fields[1]
		speed, err := strconv.Atoi(fields[2])
		if err != nil || !isValidSerialSpeed(speed) {
			return "ERROR Invalid speed"
		}
		mode := &serial.Mode{BaudRate: speed}
		handle, err := serial.Open(port, mode)
		if err != nil {
			return fmt.Sprintf("ERROR opening serial: %v", err)
		}
		state.SerialPort = port
		state.SerialSpeed = speed
		state.SerialHandle = handle
		state.Connected = true
		return "OK"
	case "DISCONNECT":
		if !state.Connected {
			return "OK"
		}
		if state.SerialHandle != nil {
			state.SerialHandle.Close()
			state.SerialHandle = nil
		}
		state.Connected = false
		return "OK"
	case "SET":
		if !state.Connected {
			return "ROBOT NOT CONNECTED"
		}
		if len(fields) < 3 {
			return "INVALID NUMBER OF PARAMETERS"
		}
		if fields[1] == "ALL" {
			resp := setAllMotors(fields[2:], state)
			if !state.Sim && state.SerialHandle != nil && resp == "OK" {
				cmd := fmt.Sprintf("SET ALL %d %d %d %d %d %d %d\n", state.Base.Position, state.Shoulder.Position, state.Elbow.Position, state.Wrist.Position, state.WristRotate.Position, state.Gripper.Position, state.Delay)
				state.SerialHandle.Write([]byte(cmd))
			}
			return resp
		} else if fields[1] == "DELAY" {
			return setDelay(fields[2:], state)
		} else if strings.HasPrefix(fields[1], "M") {
			resp := setMotor(fields[1:], state)
			if !state.Sim && state.SerialHandle != nil && resp == "OK" {
				cmd := fmt.Sprintf("SET ALL %d %d %d %d %d %d %d\n", state.Base.Position, state.Shoulder.Position, state.Elbow.Position, state.Wrist.Position, state.WristRotate.Position, state.Gripper.Position, state.Delay)
				state.SerialHandle.Write([]byte(cmd))
			}
			return resp
		}
		return "INVALID COMMAND SYNTAX"
	case "GET":
		if len(fields) == 2 && fields[1] == "STATUS" {
			if !state.Sim && state.SerialHandle != nil {
				state.SerialHandle.Write([]byte("GET STATUS\n"))
				buf := make([]byte, 128)
				n, err := state.SerialHandle.Read(buf)
				if err == nil && n > 0 {
					return strings.TrimSpace(string(buf[:n]))
				}
				return "ERROR reading status"
			}
			return state.Status()
		}
		if len(fields) == 2 && fields[1] == "LIMITS" {
			return getLimits(state)
		}
		return "INVALID COMMAND SYNTAX"
	case "MOVE":
		if !state.Connected {
			return "ROBOT NOT CONNECTED"
		}
		if len(fields) == 2 && fields[1] == "SAFETY" {
			resp := moveSafety(state)
			if !state.Sim && state.SerialHandle != nil && resp == "OK" {
				cmd := fmt.Sprintf("SET ALL %s 20\n", SafetyPositions)
				state.SerialHandle.Write([]byte(cmd))
			}
			return resp
		}
		return "INVALID COMMAND SYNTAX"
	case "PING":
		if state.Sim {
			return "DRY"
		}
		if !state.Connected {
			return "ERROR"
		}
		if state.SerialHandle != nil {
			state.SerialHandle.Write([]byte("PING\n"))
			buf := make([]byte, 32)
			n, err := state.SerialHandle.Read(buf)
			if err == nil && n > 0 {
				return strings.TrimSpace(string(buf[:n]))
			}
			return "ERROR reading ping"
		}
		return "OK"
	case "HELP":
		s := `
SET Mx <value>
SET ALL <M1> <M2> <M3> <M4> <M5> <M6> <delay>
SET DELAY <delay>
GET STATUS
GET LIMITS
MOVE SAFETY
PING
CONNECT <port> <speed>
DISCONNECT
HELP
EXIT`
		return s
	case "EXIT":
		return "BYE"
	default:
		return "UNKNOWN COMMAND"
	}
}

func setMotor(args []string, state *RobotState) string {
	if len(args) != 2 {
		return "INVALID NUMBER OF PARAMETERS"
	}
	motorIdx, err := strconv.Atoi(strings.TrimPrefix(args[0], "M"))
	if err != nil || motorIdx < 1 || motorIdx > 6 {
		return "INVALID MOTOR"
	}
	val, err := strconv.Atoi(args[1])
	if err != nil {
		return "INVALID VALUE"
	}
	motor := getMotorByIndex(state, motorIdx)
	if motor == nil {
		return "INVALID MOTOR"
	}
	if val < motor.Min || val > motor.Max {
		return fmt.Sprintf("MOTOR %s OUT OF RANGE", motorIdx)
	}
	motor.Position = val
	return "OK"
}

func setAllMotors(args []string, state *RobotState) string {
	if len(args) < 6 || len(args) > 7 {
		return "INVALID NUMBER OF PARAMETERS"
	}
	motors := []*Motor{&state.Base, &state.Shoulder, &state.Elbow, &state.Wrist, &state.WristRotate, &state.Gripper}
	for i := 0; i < 6; i++ {
		val, err := strconv.Atoi(args[i])
		if err != nil {
			return "INVALID VALUE"
		}
		if val < motors[i].Min || val > motors[i].Max {
			return "MOTOR OUT OF RANGE"
		}
		motors[i].Position = val
	}
	if len(args) == 7 {
		delay, err := strconv.Atoi(args[6])
		if err != nil || delay < 10 || delay > 30 {
			return "INVALID DELAY"
		}
		state.Delay = delay
	}
	return "OK"
}

func setDelay(args []string, state *RobotState) string {
	if len(args) != 1 {
		return "INVALID NUMBER OF PARAMETERS"
	}
	delay, err := strconv.Atoi(args[0])
	if err != nil || delay < 10 || delay > 30 {
		return "INVALID DELAY"
	}
	state.Delay = delay
	return "OK"
}

func moveSafety(state *RobotState) string {
	positions := strings.Fields(SafetyPositions)
	motors := []*Motor{&state.Base, &state.Shoulder, &state.Elbow, &state.Wrist, &state.WristRotate, &state.Gripper}
	for i := 0; i < 6; i++ {
		val, _ := strconv.Atoi(positions[i])
		motors[i].Position = val
	}
	state.Delay = 20
	return "OK"
}

func getLimits(state *RobotState) string {
	motors := []*Motor{&state.Base, &state.Shoulder, &state.Elbow, &state.Wrist, &state.WristRotate, &state.Gripper}
	limits := "LIMITS"
	for i, m := range motors {
		limits += fmt.Sprintf(" M%d:[%d,%d]", i+1, m.Min, m.Max)
	}
	limits += fmt.Sprintf(" DELAY:[%d,%d]", DelayMin, DelayMax)
	return limits
}

func getMotorByIndex(state *RobotState, idx int) *Motor {
	switch idx {
	case 1:
		return &state.Base
	case 2:
		return &state.Shoulder
	case 3:
		return &state.Elbow
	case 4:
		return &state.Wrist
	case 5:
		return &state.WristRotate
	case 6:
		return &state.Gripper
	default:
		return nil
	}
}
