#include <Braccio.h>
#include <Servo.h>
#include <InverseK.h>


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

// Setup the lengths and rotation limits for each link
Link baseLink, upperarm, forearm, hand;
// Kinematic angles for
float a0_base, a1_shoulder, a2_elbow, a3_wrist_ver;


// Quick conversion from the Braccio angle system to radians
float b2a(float b) {
  return b / 180.0 * PI - HALF_PI;
}

// Quick conversion from radians to the Braccio angle system
float a2b(float a) {
  return (a + HALF_PI) * 180 / PI;
}

bool calculateMovementXYZ(float x, float y, float z) {
  // InverseK.solve() return true if it could find a solution and false if not.

  // Calculates the angles without considering a specific approach angle
  // InverseK.solve(x, y, z, a0_base, a1_shoulder, a2_elbow, a3_wrist_ver)
  if (InverseK.solve(x, y, z, a0_base, a1_shoulder, a2_elbow, a3_wrist_ver)) {
    Serial.print('LOG: IK angles: ');
    Serial.print(a2b(a0_base));
    Serial.print(' ');
    Serial.print(a2b(a1_shoulder));
    Serial.print(' ');
    Serial.print(a2b(a2_elbow));
    Serial.print(' ');
    Serial.println(a2b(a3_wrist_ver));
  } else {
    Serial.println("LOG: No solution found!");
    return false;
  }
  return true;
}



bool calculateMovementXYZA(float x, float y, float z, float attackAngle) {

  // Calculates the angles considering a specific approach angle
  // InverseK.solve(x, y, z, a0_base, a1_shoulder, a2_elbow, a3_wrist_ver, phi)
  if (InverseK.solve(x, y, z, a0_base, a1_shoulder, a2_elbow, a3_wrist_ver, b2a(attackAngle))) {
    Serial.print('LOG: IK angles: ');
    Serial.print(a2b(a0_base));
    Serial.print(' ');
    Serial.print(a2b(a1_shoulder));
    Serial.print(' ');
    Serial.print(a2b(a2_elbow));
    Serial.print(' ');
    Serial.println(a2b(a3_wrist_ver));
  } else {
    Serial.println("LOG: No solution found!");
    return false;
  }
  return true;
}

void setup() {
  Serial.begin(115200);  // Initialize serial port for communication
  Serial.println("LOG: Available commands:");
  Serial.println("LOG: SET ALL <M1> <M2> <M3> <M4> <M5> <M6> <DELAY>");
  Serial.println("LOG SET M<x> <value>");
  Serial.println("LOG: MOVE IK <X> <Y> <Z> <ATTACKANGLE>");
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
  Braccio.ServoMovement(10, 90, 45, 180, 180, 90, 10);
  // Define phisical specs of the robot measuring from the rotation axis
  baseLink.init(72, b2a(0.0), b2a(180.0));
  upperarm.init(125, b2a(15.0), b2a(165.0));
  forearm.init(125, b2a(0.0), b2a(180.0));
  hand.init(134, b2a(0.0), b2a(180.0));

  // Attach the links to the inverse kinematic model
  InverseK.attach(baseLink, upperarm, forearm, hand);
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
      int start = 8;  // after "SET ALL "
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
      Serial.print("M1:");
      Serial.print(params[0]);
      Serial.print(" M2:");
      Serial.print(params[1]);
      Serial.print(" M3:");
      Serial.print(params[2]);
      Serial.print(" M4:");
      Serial.print(params[3]);
      Serial.print(" M5:");
      Serial.print(params[4]);
      Serial.print(" M6:");
      Serial.print(params[5]);
      Serial.print(" DELAY:");
      Serial.println(params[6]);
      // Validate ranges with detailed error reporting
      if (params[0] < 0 || params[0] > 180) {
        Serial.print("ERROR: M1 out of range: ");
        Serial.println(params[0]);
        return;
      }
      if (params[1] < 15 || params[1] > 165) {
        Serial.print("ERROR: M2 out of range: ");
        Serial.println(params[1]);
        return;
      }
      if (params[2] < 0 || params[2] > 180) {
        Serial.print("ERROR: M3 out of range: ");
        Serial.println(params[2]);
        return;
      }
      if (params[3] < 0 || params[3] > 180) {
        Serial.print("ERROR: M4 out of range: ");
        Serial.println(params[3]);
        return;
      }
      if (params[4] < 0 || params[4] > 180) {
        Serial.print("ERROR: M5 out of range: ");
        Serial.println(params[4]);
        return;
      }
      if (params[5] < 10 || params[5] > 73) {
        Serial.print("ERROR: M6 out of range: ");
        Serial.println(params[5]);
        return;
      }
      if (params[6] < 10 || params[6] > 30) {
        Serial.print("ERROR: DELAY out of range: ");
        Serial.println(params[6]);
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
    } else if (cmd.startsWith("SET M")) {
      // SET Mx <value>
      int motorIdx = cmd.charAt(5) - '0';
      int spaceIdx = cmd.indexOf(' ', 6);
      if (motorIdx < 1 || motorIdx > 6 || spaceIdx == -1) {
        Serial.println("ERROR: Invalid SET Mx command");
        return;
      }
      String valStr = cmd.substring(spaceIdx + 1);
      int value = valStr.toInt();
      bool valid = true;
      switch (motorIdx) {
        case 1:
          if (value < 0 || value > 180) { Serial.print("ERROR: M1 out of range: "); Serial.println(value); valid = false; }
          else pos_base = value;
          break;
        case 2:
          if (value < 15 || value > 165) { Serial.print("ERROR: M2 out of range: "); Serial.println(value); valid = false; }
          else pos_shoulder = value;
          break;
        case 3:
          if (value < 0 || value > 180) { Serial.print("ERROR: M3 out of range: "); Serial.println(value); valid = false; }
          else pos_elbow = value;
          break;
        case 4:
          if (value < 0 || value > 180) { Serial.print("ERROR: M4 out of range: "); Serial.println(value); valid = false; }
          else pos_wrist_ver = value;
          break;
        case 5:
          if (value < 0 || value > 180) { Serial.print("ERROR: M5 out of range: "); Serial.println(value); valid = false; }
          else pos_wrist_rot = value;
          break;
        case 6:
          if (value < 10 || value > 73) { Serial.print("ERROR: M6 out of range: "); Serial.println(value); valid = false; }
          else pos_gripper = value;
          break;
      }
      if (!valid) return;
      // LOG the received value
      Serial.print("LOG: SET M"); Serial.print(motorIdx); Serial.print(" received: "); Serial.println(value);
      Braccio.ServoMovement(pos_delay, pos_base, pos_shoulder, pos_elbow, pos_wrist_ver, pos_wrist_rot, pos_gripper);
      Serial.println("OK");
    } else if (cmd.startsWith("MOVE IK")) {
      int params[4];
      int idx = 0;
      int last = 4;
      int start = 8;  // after "MOVE IK "
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
      Serial.print("LOG: MOVE IK received: ");
      Serial.print("X:");
      Serial.print(params[0]);
      Serial.print(" Y:");
      Serial.print(params[1]);
      Serial.print(" Z:");
      Serial.print(params[2]);
      Serial.print(" ATTACK:");
      Serial.println(params[3]);

      bool angleAttackDefined = !(params[3] < 0 || params[3] > 180);

      float x = params[0];
      float y = params[1];
      float z = params[2];
      float angleAttack = params[3];
      bool IKMovementPossible = false;
      if (angleAttackDefined) {
        IKMovementPossible = calculateMovementXYZA(x, y, z, angleAttack);
      } else IKMovementPossible = calculateMovementXYZ(x, y, z);
      if (IKMovementPossible) {
      // a0_base, a1_shoulder, a2_elbow, a3_wrist_ver
        pos_base = (int)a2b(a0_base);
        pos_shoulder = (int)a2b(a1_shoulder);
        pos_elbow = (int)a2b(a2_elbow);
        pos_wrist_ver = (int)a2b(a3_wrist_ver);
        Braccio.ServoMovement(pos_delay, pos_base, pos_shoulder, pos_elbow, pos_wrist_ver, pos_wrist_rot, pos_gripper);
        Serial.println("OK");
      } else {
        Serial.println("ERROR: IK movement not possible");
      }

    } else if (cmd == "GET STATUS") {
      Serial.print("STATUS ");
      Serial.print("M1:");
      Serial.print(pos_base);
      Serial.print(" M2:");
      Serial.print(pos_shoulder);
      Serial.print(" M3:");
      Serial.print(pos_elbow);
      Serial.print(" M4:");
      Serial.print(pos_wrist_ver);
      Serial.print(" M5:");
      Serial.print(pos_wrist_rot);
      Serial.print(" M6:");
      Serial.print(pos_gripper);
      Serial.print(" DELAY:");
      Serial.print(pos_delay);
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
