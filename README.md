# Tinkerkit Braccio repository
This repository contains different tools to control and operate Tinkerkit Braccio, an arduino-powered hobbyist robotic arm.

## about Braccio
Braccio is an arduino-powered tabletop rotobic arm. Powered by 6 servo motors and controlled using any arduino board that has an arduino UNO form factor.

## repo content
- `braccio-arduino-serial-interface` contents the arduino code. Send `HELP` using the serial port to check on the commands available. 
- `braccio-network-interface` is a network proxy for the `braccio-arduino-serial-interface`. 
  - It exposes a TCP socket in port 8999 and interacts with the robot arm's arduino board using the serial port.
  - It includes a simulation mode where it doesn't need a phisical arm to react and will track the positions and delays like a real arm. 
  - The commands on the network side are compatible with the serial port ones. check the README.md for all the details.
- `braccio-inverse-kinematics` is a tool to compute the angles required for a given position in the space using the quaternions of the Braccio robot (according to [this project](https://github.com/ohlr/braccio_arduino_ros_rviz) and [this project](https://github.com/jonabalzer/braccio_description/tree/master)). 
