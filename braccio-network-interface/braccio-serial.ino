#include <Braccio.h>
#include <Servo.h>


Servo base;
Servo shoulder;
Servo elbow;
Servo wrist_rot;
Servo wrist_ver;
Servo gripper;

int pos_base = 90;
int pos_shoulder = 45;
int pos_elbow = 180;
int pos_wrist_ver = 180;
int pos_wrist_rot = 90;
int pos_gripper = 10;
int pos_delay = 20;

void setup() {
  Serial.begin(115200); // Initialize serial port for communication
  Serial.println("LOG: Available commands:");
  Serial.println("LOG: SET ALL <M1> <M2> <M3> <M4> <M5> <M6> <DELAY>");
  Serial.println("LOG: GET STATUS");
  Serial.println("LOG: PING");
  Serial.println("LOG: MOVE SAFETY");
  //Initialization functions and set up the initial position for Braccio
  //All the servo motors will be positioned in the "safety" position:
  //Base (M1):90 degrees
  //Shoulder (M2): 45 degrees
  //Elbow (M3): 180 degrees
  //Wrist vertical (M4): 180 degrees
  //Wrist rotation (M5): 90 degrees
  //gripper (M6): 10 degrees
  Braccio.begin(-70);
}

  /*
   Step Delay: a milliseconds delay between the movement of each servo.  Allowed values from 10 to 30 msec.
   M1=base degrees. Allowed values from 0 to 180 degrees
   M2=shoulder degrees. Allowed values from 15 to 165 degrees
   M3=elbow degrees. Allowed values from 0 to 180 degrees
   M4=wrist vertical degrees. Allowed values from 0 to 180 degrees
   M5=wrist rotation degrees. Allowed values from 0 to 180 degrees
   M6=gripper degrees. Allowed values from 10 to 73 degrees. 10: the toungue is open, 73: the gripper is closed.
  */

void loop() {
  if (Serial.available()) {
    String cmd = Serial.readStringUntil('\n');
    cmd.trim();
    if (cmd.startsWith("SET ALL")) {
      int params[7];
      int idx = 0;
      int last = 7;
      int start = 8; // after "SET ALL "
      for (int i = 0; i < last; i++) {
        int spaceIdx = cmd.indexOf(' ', start);
        if (spaceIdx == -1 && i < last - 1) {
          Serial.println("ERROR: Missing parameters");
          return;
        }
        String valStr = (spaceIdx == -1) ? cmd.substring(start) : cmd.substring(start, spaceIdx);
        params[i] = valStr.toInt();
        start = spaceIdx + 1;
      }
      // LOG the received values
      Serial.print("LOG: SET ALL received: ");
      Serial.print("M1:"); Serial.print(params[0]);
      Serial.print(" M2:"); Serial.print(params[1]);
      Serial.print(" M3:"); Serial.print(params[2]);
      Serial.print(" M4:"); Serial.print(params[3]);
      Serial.print(" M5:"); Serial.print(params[4]);
      Serial.print(" M6:"); Serial.print(params[5]);
      Serial.print(" DELAY:"); Serial.println(params[6]);
      // Validate ranges with detailed error reporting
      if (params[0] < 0 || params[0] > 180) {
        Serial.print("ERROR: M1 out of range: "); Serial.println(params[0]);
        return;
      }
      if (params[1] < 15 || params[1] > 165) {
        Serial.print("ERROR: M2 out of range: "); Serial.println(params[1]);
        return;
      }
      if (params[2] < 0 || params[2] > 180) {
        Serial.print("ERROR: M3 out of range: "); Serial.println(params[2]);
        return;
      }
      if (params[3] < 0 || params[3] > 180) {
        Serial.print("ERROR: M4 out of range: "); Serial.println(params[3]);
        return;
      }
      if (params[4] < 0 || params[4] > 180) {
        Serial.print("ERROR: M5 out of range: "); Serial.println(params[4]);
        return;
      }
      if (params[5] < 10 || params[5] > 73) {
        Serial.print("ERROR: M6 out of range: "); Serial.println(params[5]);
        return;
      }
      if (params[6] < 10 || params[6] > 30) {
        Serial.print("ERROR: DELAY out of range: "); Serial.println(params[6]);
        return;
      }
      pos_base = params[0];
      pos_shoulder = params[1];
      pos_elbow = params[2];
      pos_wrist_ver = params[3];
      pos_wrist_rot = params[4];
      pos_gripper = params[5];
      pos_delay = params[6];
      Braccio.ServoMovement(pos_delay, pos_base, pos_shoulder, pos_elbow, pos_wrist_ver, pos_wrist_rot, pos_gripper);
      Serial.println("OK");
    } else if (cmd == "GET STATUS") {
      Serial.print("STATUS ");
      Serial.print("M1:"); Serial.print(pos_base);
      Serial.print(" M2:"); Serial.print(pos_shoulder);
      Serial.print(" M3:"); Serial.print(pos_elbow);
      Serial.print(" M4:"); Serial.print(pos_wrist_ver);
      Serial.print(" M5:"); Serial.print(pos_wrist_rot);
      Serial.print(" M6:"); Serial.print(pos_gripper);
      Serial.print(" DELAY:"); Serial.print(pos_delay);
      Serial.println();
    } else if (cmd == "PING") {
      Serial.println("OK");
    } else if (cmd.startsWith("MOVE SAFETY")) {
      pos_delay = 20;
      pos_base = 90;
      pos_shoulder = 45;
      pos_elbow = 180;
      pos_wrist_ver = 180;
      pos_wrist_rot = 90;
      pos_gripper = 10;
      Braccio.ServoMovement(pos_delay, pos_base, pos_shoulder, pos_elbow, pos_wrist_ver, pos_wrist_rot, pos_gripper);
      Serial.println("OK");
    } else {
      Serial.println("ERROR: Unknown command");
    }
  }
}
