package bracciorobot

type Motor struct {
	Name     string
	Min      int
	Max      int
	Position int
	MsPerDeg float64 // milliseconds per degree, only used in simulation
}

type RobotState struct {
	Base        Motor
	Shoulder    Motor
	Elbow       Motor
	Wrist       Motor
	WristRotate Motor
	Gripper     Motor
	Delay       int
}
