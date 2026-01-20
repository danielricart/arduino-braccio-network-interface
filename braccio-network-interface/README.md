# Tinkerkit Braccio network-serial interface

This project provides a network-serial interface for the Tinkerkit Braccio robotic arm. It allows you to control the Braccio over a network connection using simple commands. Commands are then either printed to the console or sent via serial to the Braccio.

It includes a simulation mode that does not try to connect to the Braccio, useful for testing and development without the physical hardware.

The board is directly connected to a computer via USB Serial, and the network-serial interface runs on the computer, allowing remote control of the Braccio over the network via the computer/server.

## Features
- Control the Tinkerkit Braccio robotic arm over a network connection.
- Send commands via serial to the Braccio.
- Simple command interface for easy integration with other systems.
- Configurable network and serial settings.
- Logging capabilities for monitoring and debugging.
- tracks motor positions and reports them over the network.
- Inverse Kinematics calculations for positioning the arm in 3D space.
- Simulation mode for testing without physical hardware.

## Requirements
- Braccio robotic arm (not really needed)
- Arduino board (Uno, Mega, etc.). If it has the UNO form factor/pinout it will plug directly to the Braccio board.
- Network connectivity (Ethernet or Wi-Fi). Also works on localhost
- Arduino IDE for uploading the code to the board
- golang 1.25 or higher for building the network-serial interface. check https://go.dev/dl/ on how to install golang for your specific platform.
- Tinkerkit Braccio library

## Installation
The project is composed of two parts. The robot side (Arduino) and the network-serial interface side (Golang). 

Clone this repository to your local machine.

### Robot side (Arduino)
1. Open the Arduino IDE 
2. install the Tinkerkit Braccio library. You can do this via the Library Manager (Sketch -> Include Library -> Manage Libraries...) and searching for "Braccio". Install "Braccio by Andrea Martino, Arduino".
3. Install the CGx-InverseK library for Inverse Kinematics calculations. You can find it here: https://github.com/cgxeiji/CGx-InverseK 
4. Connect your board to the computer. 
5. Select your boaard from the boards list. Upload the provided Arduino sketch braccio-arduino-serial-interface.ino to your Arduino board.
6. Connect the Braccio robotic arm to the Arduino board.
7. Connect the Arduino board to your computer via USB.
8. Open the serial monitor port and type `PING` to check if the robot is responding. You should see `OK` if everything is working.


### Network-serial interface side (Golang)
1. Make sure you have golang 1.25 or higher installed on your machine. Use the command `go version` to check your golang version.
2. If you don't have golang installed, follow the instructions on https://go.dev/dl/ to install it for your platform.
3. Open a terminal and navigate to the cloned repository.
4. Build the network-serial interface using the command:
   ```bash
   go build -o braccio-network-serial-interface main.go
   ```
5. Run the network-serial interface with the command:
   ```bash
   ./braccio-network-serial-interface
   ```
6. Establish the connection to the board using the `CONNECT <port> <speed>` command, where `<port>` is the serial port of your Arduino board (e.g., `/dev/ttyUSB0` on Linux/Mac or `COM3` on Windows) and `<speed>` is the baud rate (Arduino code is set to `115200`).

## Braccio specifications
### Motor limits
- M1=base degrees. Allowed values from 0 to 180 degrees 
- M2=shoulder degrees. Allowed values from 15 to 165 degrees 
- M3=elbow degrees. Allowed values from 0 to 180 degrees 
- M4=wrist vertical degrees. Allowed values from 0 to 180 degrees 
- M5=wrist rotation degrees. Allowed values from 0 to 180 degrees 
- M6=gripper degrees. Allowed values from 10 to 73 degrees. 10: the gripper is open, 73: the gripper is closed.
- Propagation time between motor movements: 10 to 30 ms. configurable

### Motor speed
- M1: speed 3.3 milliseconds per degree
- M2: speed 3.3 milliseconds per degree
- M3: speed 3.3 milliseconds per degree
- M4: speed 3.3 milliseconds per degree
- M5: speed 2.3 milliseconds per degree
- M6: speed 2.3 milliseconds per degree

### Safety position
- Base (M1):90 degrees
- Shoulder (M2): 45 degrees
- Elbow (M3): 180 degrees
- Wrist vertical (M4): 180 degrees
- Wrist rotation (M5): 90 degrees
- gripper (M6): 10 degrees
- Propagation time between motor movements: 20 ms

### Articulations and arms lengths
- Base articulation height: 72 mm
- Arm length: 125 mm
- Forearm length: 55 mm
- Hand length: 134 mm
- Maximum reach: 525 mm

## Protocol definition
### Network side
The network side of the communications allow sending motor positions and receiving robot status (motor positions). The protocol is line-based, with each command or response terminated by a newline character (`\n`).

There are commands to set each motor position individually or all motors at once, as well as a command to get the current status of the robot, and a command to move the robot to Safety position.

#### Commands
- `SET Mx <value>`: Set selected motor `x` (x=[1, 6])  to `<value>` degrees. `<value>` must be within the allowed range for the motor.
- `SET ALL <M1> <M2> <M3> <M4> <M5> <M6> <delay>`: Set all motors to the specified positions. using the specified propagation `delay` in milliseconds. `<delay>` is optional; if not provided, the default delay will be used.
- `SET DELAY <delay>`: Set the default propagation delay between motor movements to `<delay>` milliseconds. This delay will be used for subsequent motor movements unless overridden by a specific command.
- `GET STATUS`: Get the current status of all motors and delay. The response will be in the format: `STATUS M1:<value> M2:<value> M3:<value> M4:<value> M5:<value> M6:<value> DELAY:<value>`.
- `GET LIMITS`: Get the allowed limits for each motor. The response will be in the format: `LIMITS M1:<min>-<max> M2:<min>-<max> M3:<min>-<max> M4:<min>-<max> M5:<min>-<max> M6:<min>-<max> DELAY:<min>-<max>`.
- `MOVE SAFETY`: Move all motors to the predefined safety position. Always use 20ms delay.
- `PING`: Check if the connection is alive. The response will be `OK` if the robot is available. `DRY` if working on simulation mode. `ERROR` if the robot is not available due to any error.
- `CONNECT <port> <speed>`: Connect to the Braccio robot on the specified serial `<port>`. If already connected, it will first disconnect and then connect to the new port and speed. <speed> will set the serial speed (one of 9600, 19200, 38400, 57600, 115200, 230400, 460800, 921600). It will respond `OK` if successful, or `ERROR <message>` if there was an error.
- `DISCONNECT`: Disconnect from the Braccio robot. It will respond `OK`. If already disconnected, it will respond with `OK`.
- `HELP`: Get a list of available commands.
- `EXIT`: Close the network connection.

#### Responses
- `OK`: Command executed successfully.
- `ERROR <message>`: An error occurred while processing the command. `<message>` provides details about the error.
- `STATUS M1:<value> M2:<value> M3:<value> M4:<value> M5:<value> M6:<value> DELAY:<value>`: Current status of all motors and current delay.
- `DRY`: The robot is in simulation mode and not connected to the physical hardware.
- `HELP <commands>`: List of available commands.
- `BYE`: Connection closed.
- `UNKNOWN COMMAND`: The command received is not recognized.
- `INVALID PARAMETERS`: The parameters provided with the command are not valid.
- `MOTOR OUT OF RANGE`: The specified motor value is outside the allowed range.
- `ROBOT NOT AVAILABLE`: The robot is not available due to a connection or hardware error.
- `ROBOT BUSY`: The robot is currently executing another command and cannot process the new command.
- `INVALID DELAY`: The specified delay value is not valid.
- `INVALID MOTOR`: The specified motor number is not valid.
- `INVALID FORMAT`: The command format is incorrect.
- `INVALID VALUE`: The specified value is not valid.
- `INVALID NUMBER OF PARAMETERS`: The number of parameters provided with the command is incorrect.
- `INVALID COMMAND SYNTAX`: The syntax of the command is incorrect.
- `INVALID COMMAND FORMAT`: The format of the command is incorrect.

### Serial side
The serial side of the communications is used to send commands to the Braccio robotic arm. The commands send all motor positions and a delay between motor movements. It can also send a status request to get the current motor positions and robot status. 

Serial communications will be set at 115200 bps to minimize latency and blocking time.

The library used to communicate with the serial port is go.bug.st/serial. Documentation is here: https://pkg.go.dev/go.bug.st/serial

#### Commands
- `SET ALL <M1> <M2> <M3> <M4> <M5> <M6> <delay>`: Set all motors to the specified positions. using the specified propagation `delay` in milliseconds.
- `SET Mx <value>`: Set selected motor `x` (x=[1, 6])  to `<value>` degrees. `<value>` must be within the allowed range for the motor.
- `GET STATUS`: Get the current status of all motors and delay. The response will be in the format: `STATUS M1:<value> M2:<value> M3:<value> M4:<value> M5:<value> M6:<value> DELAY:<value>`.
- `MOVE SAFETY`: Move all motors to the predefined safety position. Always use 20ms delay.
- `PING`: Check if the connection is alive. The response will be `OK` if the robot is available. `ERROR` if the robot is not available due to any error.

#### Responses
- `OK`: Command executed successfully.
- `ERROR <message>`: An error occurred while processing the command. `<message>` provides details about the error.
- `STATUS M1:<value> M2:<value> M3:<value> M4:<value> M5:<value> M6:<value> DELAY:<value>`: Current status of all motors and current delay.
- `UNKNOWN COMMAND`: The command received is not recognized.
- `INVALID PARAMETERS`: The parameters provided with the command are not valid.
- `MOTOR OUT OF RANGE`: The specified motor value is outside the allowed range.
- `ROBOT NOT AVAILABLE`: The robot is not available due to a connection or hardware error
- `ROBOT BUSY`: The robot is currently executing another command and cannot process the new command.
- `INVALID DELAY`: The specified delay value is not valid.
- `INVALID MOTOR`: The specified motor number is not valid.
- `INVALID FORMAT`: The command format is incorrect.
- `INVALID VALUE`: The specified value is not valid.

# references
This section is for reference materials related to the Tinkerkit Braccio and the libraries used in this project. It's not used directly in the code but provides useful information for understanding and working with the Braccio robotic arm.

- Tinkerkit Braccio official website: https://www.arduino.cc/en/Tinkerkit/Braccio
- Tinkerkit Braccio library: https://github.com/arduino-libraries/Braccio/tree/master
- go.bug.st/serial library: https://pkg.go.dev/go.bug.st/serial
- Inverse Kinematics for arduino https://github.com/cgxeiji/CGx-InverseK
- Alternative Inverse Kinematics for arduino: https://github.com/henriksod/Fabrik2DArduino 
- Braccio URDF model: https://github.com/jonabalzer/braccio_description/tree/master
- Braccio ROS package and alternative URDF: https://github.com/ohlr/braccio_arduino_ros_rviz

# Future improvements
## Arduino Q
There's this new Arduino Q boards that have both a computer side (Zephyr OS Linux) and a microcontroller side (Arduino). It would be interesting to port the network-serial interface to run directly on the Arduino Q board, eliminating the need for an external computer and making the arm fully standalone, provided it can be connected to the network via Wi-Fi.

A possible implementation would be to migrate from the Serial Port communications to the Arduino Q's native built-in RPC library Arduino Bridge ([Docs here](https://docs.arduino.cc/tutorials/uno-q/user-manual/#communication), but probably other approaches can be taken depending on the state-of-the-art and coding styles and preferences. 

## Better Inverse Kinematics
THe current implementation for IK is clumsy and cannot compute scenarios that are probably reachable. Other libraries like Fabrik2DArduino may work a bit better.
An alternative to this would be to translate the IK calculations to a computer-based model that only transfers the angular positions to the robot. This can be integrating the provided external tool for IK calculations `braccio-inverse-kinematics` or a new one.

## 3D Simulator
A robotic arm is a physical piece in the real world. Sometimes it's not possible to have it connected. When exploring movements and integrations it can even be dangerously destructive for the real piece or its surroundings to use the real arm. Having a 3D representation of it that can connect to the network-serial interface and simulate the movements of the arm would be very useful. The network-serial interface already has a dry-run mode that can be used for this purpose.

There are links to the UDRF model of the Braccio in the references section that can be used as a starting point for this simulator, as they even contain some 3D assets that can be used for this. 

# Contributing
Contributions are welcome! If you find a bug or have a feature request, please open an issue on GitHub. If you want to contribute code, please fork the repository and create a pull request.

If you build upon this project, please give appropriate credit, and provide a link to the original project. You may use this project for commercial purposes as long as you comply with the license terms.
