# Real-time elevator project

This is a project where three elevators are connected via UDP to communicate and distribute orders

## How to run

A script is added with some hard-coded parameters and ports.
To run this:

```bash
chmod +x run_elevs.sh
./run_elevs.sh
```

## Simulator
The simulator folder is provided, but the simulator executable can be found here:
https://github.com/TTK4145/Simulator-v2

## Modules

### FSM

Our FSM module is designed to take care of the elevator itself. It consists of a main Fsm function which is launched as it's own goroutine and handles the behavior of the elevator itself, and an internalControl function which communicates with the rest of the modules and the elevator itself.

### config

Config contains the different constants and structs that are used by the other modules. This was done to keep all the different defenitions within one file, making them easy to find and edit. Unless only used for a spesific module everything is defined here.

### elevNet

elevNet contains two functions, one for sending and one for recieving. They handle communication with the other elevators, through the network module. The elevators all send information on themselves over UDP to the other elevators, where it is recieved and stored. They also send and recieve any order which is sent between the elevators.

### network

We are using the network module provided, which uses UDP to send and recieve information.

### orderManager

The orderManager takes care of delegating orders. It has information about the current state of all the elevators, and delegates orders using a simple cost function. If an elevator is lost it is also responsible for transfering the orders to any of the remaining operational elevators.
