package simrobot

import (
	"errors"
	"time"

	bracciorobot "braccio-network-interface/bracciorobot"
)

type Motor = bracciorobot.Motor

type RobotState = bracciorobot.RobotState

// SimulateMovement moves specified motors to their target positions, one degree at a time, with correct timing.
// targets: map[motorID]targetPosition, where motorID is 1-6 (M1-M6)
func SimulateMovement(state *RobotState, targets map[int]int, delay int) error {
	motors := []*Motor{&state.Base, &state.Shoulder, &state.Elbow, &state.Wrist, &state.WristRotate, &state.Gripper}

	if len(targets) == 0 {
		return errors.New("no targets specified")
	}
	for id := range targets {
		if id < 1 || id > 6 {
			return errors.New("invalid motor ID")
		}
	}

	done := false
	for !done {
		done = true
		for i, m := range motors {
			target, ok := targets[i+1]
			if !ok {
				continue // skip motors not in targets
			}
			if m.Position != target {
				done = false
				if target > m.Position {
					m.Position++
				} else {
					m.Position--
				}
				// Use msPerDeg from the struct (milliseconds per degree)
				if m.MsPerDeg > 0 {
					time.Sleep(time.Duration(m.MsPerDeg * float64(time.Millisecond)))
				}
			}
		}
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	return nil
}
