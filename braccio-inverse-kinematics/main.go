package main

import (
	"fmt"
	"github.com/koeng101/kinematics"
	"math"
)

// RadiansToDegrees converts a value from radians to degrees.
func RadiansToDegrees(radians float64) float64 {
	return radians * (180.0 / math.Pi)
}

// DegreesToRadians converts a value from degrees to radians.
func DegreesToRadians(degrees float64) float64 {
	return degrees * (math.Pi / 180.0)
}

func main() {
	// Establish the original joint angles
	thetasInit := kinematics.JointAngles{0, 0, 0, 0, 0, 0}

	// Establish coordinates to go to
	coordinates := kinematics.Pose{
		Position: kinematics.Position{X: -100, Y: 250, Z: 250},
		Rotation: kinematics.Quaternion{W: 0.41903052745255764, X: 0.4007833787652043, Y: -0.021233218878182854, Z: 0.9086418268616911},
	}

	// Run kinematics procedure
	angles, _ := kinematics.InverseKinematics(coordinates, kinematics.AR3DhParameters, thetasInit)

	// Math works slightly differently on arm and x86 machines when calculating
	// inverse kinematics. We check 5 decimals deep, since it appears numbers can
	// have slight variations between arm and x86 at 6 decimals.
	fmt.Printf("%5f, %5f, %5f, %5f, %5f, %5f\n", angles.J1, angles.J2, angles.J3, angles.J4, angles.J5, angles.J6)
}
