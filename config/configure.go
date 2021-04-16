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
	ExtOrder      chan elevio.ButtonEvent
	DelegateOrder chan elevio.ButtonEvent
	SendOrder     chan elevio.ButtonEvent
	ExternalID    chan string
}

type ElevChannels struct {
	Elevator chan Elev
}

type NetworkMessage struct {
	Elevator  Elev
	ID        string
	OrderIncl bool
	Order     elevio.ButtonEvent
}
