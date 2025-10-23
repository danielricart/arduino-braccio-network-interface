package main

import (
	bracciorobot "braccio-network-interface/bracciorobot"
	simrobot "braccio-network-interface/simrobot"
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

type Motor = bracciorobot.Motor

type RobotState struct {
	bracciorobot.RobotState
	Sim          bool
	Connected    bool
	SerialPort   string
	SerialSpeed  int
	SerialHandle serial.Port
}

func NewRobotState(sim bool) *RobotState {
	rs := &RobotState{
		RobotState: bracciorobot.RobotState{
			Base:        bracciorobot.Motor{"Base", 0, 180, 90, 0},
			Shoulder:    bracciorobot.Motor{"Shoulder", 15, 165, 45, 0},
			Elbow:       bracciorobot.Motor{"Elbow", 0, 180, 180, 0},
			Wrist:       bracciorobot.Motor{"Wrist", 0, 180, 180, 0},
			WristRotate: bracciorobot.Motor{"WristRotate", 0, 180, 90, 0},
			Gripper:     bracciorobot.Motor{"Gripper", 10, 73, 10, 0},
			Delay:       DefaultDelay,
		},
		Sim:       sim,
		Connected: false,
	}
	if sim {
		// Set MsPerDeg for simulation mode
		rs.Base.MsPerDeg = 3.3
		rs.Shoulder.MsPerDeg = 3.3
		rs.Elbow.MsPerDeg = 3.3
		rs.Wrist.MsPerDeg = 3.3
		rs.WristRotate.MsPerDeg = 2.3
		rs.Gripper.MsPerDeg = 2.3
	}
	return rs
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
		s, done := disconnect(state)
		if done {
			return s
		}
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
		_ = moveSafety(state)
		_, _ = disconnect(state)
		return "BYE"
	default:
		return "UNKNOWN COMMAND"
	}
}

func disconnect(state *RobotState) (string, bool) {
	if !state.Connected {
		return "OK", true
	}
	if state.SerialHandle != nil {
		err := state.SerialHandle.Close()
		if err != nil {
			return "", false
		}
		state.SerialHandle = nil
	}
	state.Connected = false
	return "", false
}

func toSimRobotState(state *RobotState) *simrobot.RobotState {
	return &simrobot.RobotState{
		Base:        simrobot.Motor(state.Base),
		Shoulder:    simrobot.Motor(state.Shoulder),
		Elbow:       simrobot.Motor(state.Elbow),
		Wrist:       simrobot.Motor(state.Wrist),
		WristRotate: simrobot.Motor(state.WristRotate),
		Gripper:     simrobot.Motor(state.Gripper),
		Delay:       state.Delay,
	}
}

func updateFromSimRobotState(state *RobotState, sim *simrobot.RobotState) {
	state.Base = bracciorobot.Motor(sim.Base)
	state.Shoulder = bracciorobot.Motor(sim.Shoulder)
	state.Elbow = bracciorobot.Motor(sim.Elbow)
	state.Wrist = bracciorobot.Motor(sim.Wrist)
	state.WristRotate = bracciorobot.Motor(sim.WristRotate)
	state.Gripper = bracciorobot.Motor(sim.Gripper)
	state.Delay = sim.Delay
}

func toSimRobotTargetsFromSlice(slice []int) map[int]int {
	targets := make(map[int]int)
	for i, v := range slice {
		targets[i+1] = v
	}
	return targets
}

func setDelay(args []string, state *RobotState) string {
	if len(args) != 1 {
		return "INVALID NUMBER OF PARAMETERS"
	}
	val, err := strconv.Atoi(args[0])
	if err != nil || val < DelayMin || val > DelayMax {
		return "INVALID DELAY"
	}
	state.Delay = val
	return "OK"
}

func setAllMotors(args []string, state *RobotState) string {
	if len(args) < 6 || len(args) > 7 {
		return "INVALID NUMBER OF PARAMETERS"
	}
	motors := []*Motor{&state.Base, &state.Shoulder, &state.Elbow, &state.Wrist, &state.WristRotate, &state.Gripper}
	targetsSlice := make([]int, 6)
	for i := 0; i < 6; i++ {
		val, err := strconv.Atoi(args[i])
		if err != nil {
			return "INVALID VALUE"
		}
		if val < motors[i].Min || val > motors[i].Max {
			return "MOTOR OUT OF RANGE"
		}
		targetsSlice[i] = val
	}
	delay := state.Delay
	if len(args) == 7 {
		d, err := strconv.Atoi(args[6])
		if err != nil || d < 10 || d > 30 {
			return "INVALID DELAY"
		}
		delay = d
	}
	if state.Sim {
		simState := toSimRobotState(state)
		targets := toSimRobotTargetsFromSlice(targetsSlice)
		err := simrobot.SimulateMovement(simState, targets, delay)
		if err != nil {
			return "SIM ERROR"
		}
		updateFromSimRobotState(state, simState)
		return "OK"
	}
	for i := 0; i < 6; i++ {
		motors[i].Position = targetsSlice[i]
	}
	if len(args) == 7 {
		state.Delay = delay
	}
	return "OK"
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
		return fmt.Sprintf("MOTOR %s OUT OF RANGE", args[0])
	}
	if state.Sim {
		targets := map[int]int{motorIdx: val}
		simState := toSimRobotState(state)
		err := simrobot.SimulateMovement(simState, targets, state.Delay)
		if err != nil {
			return "SIM ERROR"
		}
		updateFromSimRobotState(state, simState)
		return "OK"
	}
	motor.Position = val
	return "OK"
}

func moveSafety(state *RobotState) string {
	positions := strings.Fields(SafetyPositions)
	motors := []*Motor{&state.Base, &state.Shoulder, &state.Elbow, &state.Wrist, &state.WristRotate, &state.Gripper}
	targetsSlice := make([]int, 6)
	for i := 0; i < 6; i++ {
		val, _ := strconv.Atoi(positions[i])
		targetsSlice[i] = val
	}
	if state.Sim {
		simState := toSimRobotState(state)
		targets := toSimRobotTargetsFromSlice(targetsSlice)
		err := simrobot.SimulateMovement(simState, targets, 20)
		if err != nil {
			return "SIM ERROR"
		}
		updateFromSimRobotState(state, simState)
		return "OK"
	}
	for i := 0; i < 6; i++ {
		motors[i].Position = targetsSlice[i]
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
