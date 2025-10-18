package main

import (
	"bufio"
	"flag"
	"fmt"
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

type Motor struct {
	Name     string
	Min      int
	Max      int
	Position int
}

type RobotState struct {
	Base        Motor
	Shoulder    Motor
	Elbow       Motor
	Wrist       Motor
	WristRotate Motor
	Gripper     Motor
	Delay       int
	Sim         bool
	Connected   bool
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
			return "ALREADY CONNECTED"
		}
		// In real mode, would open serial port here. For now, just set Connected.
		state.Connected = true
		return "CONNECTED"
	case "DISCONNECT":
		if !state.Connected {
			return "ALREADY DISCONNECTED"
		}
		// In real mode, would close serial port here. For now, just set Connected.
		state.Connected = false
		return "DISCONNECTED"
	case "SET":
		if !state.Connected {
			return "ROBOT NOT CONNECTED"
		}
		if len(fields) < 3 {
			return "INVALID NUMBER OF PARAMETERS"
		}
		if fields[1] == "ALL" {
			return setAllMotors(fields[2:], state)
		} else if fields[1] == "DELAY" {
			return setDelay(fields[2:], state)
		} else if strings.HasPrefix(fields[1], "M") {
			return setMotor(fields[1:], state)
		}
		return "INVALID COMMAND SYNTAX"
	case "GET":
		if len(fields) == 2 && fields[1] == "STATUS" {
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
			return moveSafety(state)
		}
		return "INVALID COMMAND SYNTAX"
	case "PING":
		if state.Sim {
			return "DRY"
		}
		if !state.Connected {
			return "ERROR"
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
