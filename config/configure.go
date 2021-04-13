package config

import (
	"../driver/elevio"
)

const numElevs = 3

type DriverChannels struct {
	DrvButtons     chan elevio.ButtonEvent
	DrvFloors      chan int
	DrvStop        chan bool
	DoorsOpen      chan int
	CompletedOrder chan elevio.ButtonEvent
	DrvObstr       chan bool
}

type OrderChannels struct {
	ExtOrder       chan elevio.ButtonEvent
	DelegateOrder  chan elevio.ButtonEvent
	OthersLocation chan [numElevs]int
}
