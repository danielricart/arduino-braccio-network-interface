# Tinkerkit Braccio network-serial interface

This project provides a network-serial interface for the Tinkerkit Braccio robotic arm. It allows you to control the Braccio over a network connection using simple commands. Commands are then either printed to the console or sent via serial to the Braccio.

It includes a simulation mode that does not try to connect to the Braccio, useful for testing and development without the physical hardware.

## Features
- Control the Tinkerkit Braccio robotic arm over a network connection.
- Send commands via serial to the Braccio.
- Simple command interface for easy integration with other systems.
- Configurable network and serial settings.
- Logging capabilities for monitoring and debugging.
- tracks motor positions and reports them over the network.

## Requirements
- Braccio robotic arm
- Arduino board compatible with Braccio. Requires Serial over USB.
- Network connectivity (Ethernet or Wi-Fi). Works on localhost
- Arduino IDE for uploading the code to the board
- Tinkerkit Braccio library

## Braccio specifications
### Motor limits
- M1=base degrees. Allowed values from 0 to 180 degrees 
- M2=shoulder degrees. Allowed values from 15 to 165 degrees 
- M3=elbow degrees. Allowed values from 0 to 180 degrees 
- M4=wrist vertical degrees. Allowed values from 0 to 180 degrees 
- M5=wrist rotation degrees. Allowed values from 0 to 180 degrees 
- M6=gripper degrees. Allowed values from 10 to 73 degrees. 10: the gripper is open, 73: the gripper is closed.
- Propagation time between motor movements: 10 to 30 ms. configurable

### Safety position
- Base (M1):90 degrees
- Shoulder (M2): 45 degrees
- Elbow (M3): 180 degrees
- Wrist vertical (M4): 180 degrees
- Wrist rotation (M5): 90 degrees
- gripper (M6): 10 degrees
- Propagation time between motor movements: 20 ms

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

#### Commands
- `SET ALL <M1> <M2> <M3> <M4> <M5> <M6> <delay>`: Set all motors to the specified positions. using the specified propagation `delay` in milliseconds.
- `GET STATUS`: Get the current status of all motors and delay. The response will be in the format: `STATUS M1:<value> M2:<value> M3:<value> M4:<value> M5:<value> M6:<value> DELAY:<value>`.
- `MOVE SAFETY`: Move all motors to the predefined safety position. Always use 20ms delay.
- `PING`: Check if the connection is alive. The response will be `OK` if the robot is available. `ERROR` if the robot is not available due to any error.