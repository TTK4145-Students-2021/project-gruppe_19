package main

import (
	"flag"
	"strconv"

	"./FSM"

	"./config"
	"./driver/elevio"
	"./elevNet"
	"./network/bcast"
	"./network/peers"
	"./ordermanager"
)

func main() {
	//Flags to configure program from call in terminal
	elevPort_p := flag.String("elev_port", "15657", "The port of the elevator to connect to (for sim purposes)")
	transmitPort_p := flag.String("transmit_port", "14654", "Port to transmit to other elevator")
	receivePort_p := flag.String("receive_port", "15555", "Port to receive from other elevator")
	receivePort2_p := flag.String("receive_port2", "4321", "Second receive port")

	id_p := flag.String("elev_id", "id", "id of this peer")
	flag.Parse()

	elevPort := *elevPort_p
	receivePort := *receivePort_p
	transmitPort := *transmitPort_p
	id := *id_p
	receivePort2 := *receivePort2_p

	hostString := "localhost:" + elevPort

	println("Connecting to server")
	elevio.Init(hostString, config.NumFloors)

	//The elevator object the state machine will handle
	var elevator = config.Elev{
		State: config.IDLE,
		Dir:   config.STILL,
		Floor: 0,
		Queue: [config.NumFloors][config.NumButtons]bool{},
	}

	//List of active elevators and their last received elevator object
	activeElevators := [config.NumElevs]bool{}
	elevatorArray := [config.NumElevs]config.Elev{}

	//Making all channels needed to run program
	driverChannels := config.DriverChannels{
		DrvButtons: make(chan elevio.ButtonEvent),
		DrvFloors:  make(chan int),
		DrvStop:    make(chan bool),
		DoorsOpen:  make(chan int),
		DrvObstr:   make(chan bool),
	}

	orderChannels := config.OrderChannels{
		ExtOrder:       make(chan elevio.ButtonEvent),
		DelegateOrder:  make(chan elevio.ButtonEvent),
		SendOrder:      make(chan elevio.ButtonEvent),
		ExternalID:     make(chan string),
		LostConnection: make(chan string),
		CompletedOrder: make(chan elevio.ButtonEvent),
	}

	elevChannels := config.ElevChannels{ //doesnt need to be a struct as it stands now. TODO: remove struct
		Elevator: make(chan config.Elev),
	}

	go elevio.PollObstructionSwitch(driverChannels.DrvObstr)
	go elevio.PollButtons(driverChannels.DrvButtons)
	go elevio.PollFloorSensor(driverChannels.DrvFloors)
	go elevio.PollStopButton(driverChannels.DrvStop)

	go FSM.Fsm(elevChannels, &elevator, driverChannels)
	go ordermanager.OrderMan(orderChannels, elevChannels, id, &elevator, &activeElevators, &elevatorArray)

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	//Making all received ports into ints
	receiveInt, _ := strconv.Atoi(receivePort)
	transmitInt, _ := strconv.Atoi(transmitPort)
	receiveInt2, _ := strconv.Atoi(receivePort2)

	//Making the peer-update messages send on the message ports +1 to reduce the needed amount of flags
	go peers.Transmitter(transmitInt+1, id, peerTxEnable)
	go peers.Receiver(receiveInt+1, peerUpdateCh)
	go peers.Receiver(receiveInt2+1, peerUpdateCh)

	networkTx := make(chan config.NetworkMessage)
	networkRx := make(chan config.NetworkMessage)

	//broadcasting and receiving to/from the first elevator
	go bcast.Transmitter(transmitInt, networkTx)
	go bcast.Receiver(receiveInt, networkRx)

	//receiving from the second elevator
	go bcast.Receiver(receiveInt2, networkRx)

	//Handles parsing and handling of messages sent and received
	go elevNet.SendElev(networkTx, elevChannels, id, orderChannels, &elevator)
	go elevNet.ReceiveElev(networkRx, elevChannels, peerUpdateCh, id, orderChannels, &activeElevators, &elevatorArray)

	//LETS go!!!!!
	FSM.InternalControl(driverChannels, orderChannels, elevChannels, &elevator)

}
