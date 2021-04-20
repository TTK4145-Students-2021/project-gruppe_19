package config

import (
	"../driver/elevio"
)

const (
	NumFloors  int = 4
	NumElevs       = 3
	NumButtons     = 3
)

type ElevatorState int

const (
	IDLE       = 0
	RUNNING    = 1
	DOOR_OPEN  = 2
	ERROR      = 3
	OBSTRUCTED = 4
)

type Direction int

const (
	UP    = 1
	DOWN  = -1
	STILL = 0
)

type Elev struct {
	State ElevatorState               //Carries the current state of the elevator
	Dir   Direction                   //Carries the current directoion the elevator is going
	Floor int                         //Carries which floor the elevator is in. If between floors, this is -1
	Queue [NumFloors][NumButtons]bool //Carries the orders the elevator has
}

type DriverChannels struct {
	DrvButtons chan elevio.ButtonEvent //sends if an order button is pressed
	DrvFloors  chan int                //sends if an elevator reaches a floor sensor
	DrvStop    chan bool               //sends if the stop button is pressed
	DoorsOpen  chan int                //sends if the doors are open
	DrvObstr   chan bool               //sends if the door is obstructed
}

type OrderChannels struct {
	ExtOrder       chan elevio.ButtonEvent //order coming into FSM to be handled
	DelegateOrder  chan elevio.ButtonEvent //order coming from FSM to be delegated to correct elevator
	SendOrder      chan elevio.ButtonEvent //order coming form ordermanager to be sent to a different elevator
	ExternalID     chan string             // ID of the elevator which is going to receive order
	LostConnection chan string             //Channels to send the ID of an elevator which loses connection
	CompletedOrder chan elevio.ButtonEvent //completed orders are sent here
}

type ElevChannels struct {
	Elevator chan Elev //channel to send the current elevator object
}

type NetworkMessage struct {
	Elevator          Elev               //elevator object
	ID                string             //ID of the elevator being sent
	TakeOrder         bool               //bool to signal if an order is included or not
	Order             elevio.ButtonEvent //order
	SetOrderLight     bool               //bool to tell if the elevator should set a light at the order
	TurnOffOrderLight bool               //bool to tell if the elevator should turn off a light at the order
}
