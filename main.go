package main

import (
	"flag"
	"fmt"

	"./FSM"
	"./config"
	"./driver/elevio"
	"./ordermanager"
)

const numFloors = 4
const numButtons = 3
const numElevs = 3

func main() {
	elevPort_p := flag.String("elev_port", "15657", "The port of the elevator to connect to (for sim purposes)")
	//får noen ganger out of index error i ordersInFloor funksjonen når forskjellige porter brukes!!!??!

	flag.Parse()

	elevPort := *elevPort_p
	hostString := "localhost:" + elevPort

	fmt.Println("Elevport ", hostString)

	println("Connecting to server")
	elevio.Init(hostString, numFloors)

	driverChannels := config.DriverChannels{
		DrvButtons:     make(chan elevio.ButtonEvent),
		DrvFloors:      make(chan int),
		DrvStop:        make(chan bool),
		DoorsOpen:      make(chan int),
		CompletedOrder: make(chan elevio.ButtonEvent, 100),
		DrvObstr:       make(chan bool),
	}

	orderChannels := config.OrderChannels{
		ExtOrder:       make(chan elevio.ButtonEvent),
		DelegateOrder:  make(chan elevio.ButtonEvent),
		OthersLocation: make(chan [numElevs]int),
	}

	/*dummyString := "fuck u bitch"
	transmitEnable := make(chan bool)
	go peers.Transmitter(22349, dummyString, transmitEnable)*/

	go elevio.PollObstructionSwitch(driverChannels.DrvObstr)
	go elevio.PollButtons(driverChannels.DrvButtons)
	go elevio.PollFloorSensor(driverChannels.DrvFloors)
	go elevio.PollStopButton(driverChannels.DrvStop)
	go FSM.Fsm(driverChannels.DoorsOpen)

	go ordermanager.OrderMan(orderChannels)

	FSM.InternalControl(driverChannels, orderChannels)

}
