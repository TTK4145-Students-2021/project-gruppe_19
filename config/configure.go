package config

import (
	"../driver/elevio"
)

const numFloors = 4
const numButtons = 3

type ElevatorState int

const (
	IDLE      = 0
	RUNNING   = 1
	DOOR_OPEN = 2
	ERROR     = 3
)

type Direction int

const (
	UP    = 1
	DOWN  = -1
	STILL = 0
)

type Elev struct {
	State ElevatorState
	Dir   Direction
	Floor int
	Queue [numFloors][numButtons]bool
}

type DriverChannels struct {
	DrvButtons     chan elevio.ButtonEvent
	DrvFloors      chan int
	DrvStop        chan bool
	DoorsOpen      chan int
	CompletedOrder chan elevio.ButtonEvent
	DrvObstr       chan bool
}

type OrderChannels struct {
	ExtOrder       chan elevio.ButtonEvent //order coming into FSM to be handled
	DelegateOrder  chan elevio.ButtonEvent //order coming from FSM to be delegated to correct elevator
	SendOrder      chan elevio.ButtonEvent //order coming form ordermanager to be sent to a different elevator
	ExternalID     chan string             // ID of the elevator which is going to receive order
	LostConnection chan string
}

type ElevChannels struct {
	Elevator chan Elev
}

type NetworkMessage struct {
	Elevator  Elev               //elevator object
	ID        string             //ID of the elevator being sent
	OrderIncl bool               //bool to signal if an order is included or not
	Order     elevio.ButtonEvent //order
	Connected bool
}
