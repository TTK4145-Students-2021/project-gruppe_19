package main

import (
	"flag"
	"fmt"
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
	//numElevs_p := flag.Int("num_elevs", 3, "Number of elevators working")
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
	//numElevs := *numElevs_p
	receivePort2 := *receivePort2_p

	hostString := "localhost:" + elevPort

	fmt.Println("Elevport ", hostString)

	println("Connecting to server")
	elevio.Init(hostString, config.NumFloors)

	var elevator = config.Elev{
		State: config.IDLE,
		Dir:   config.STILL,
		Floor: 0, //denne har ingenting å si siden den oppdateres i FSMinit
		Queue: [config.NumFloors][config.NumButtons]bool{},
	}

	activeElevators := [3]bool{true, true, true}
	elevatorArray := [3]config.Elev{}

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
		SendOrder:      make(chan elevio.ButtonEvent),
		ExternalID:     make(chan string),
		LostConnection: make(chan string),
		CompletedOrder: make(chan elevio.ButtonEvent),
	}

	elevChannels := config.ElevChannels{ //doesnt need to be a struct as it stands now. TODO: remove struct
		Elevator: make(chan config.Elev),
	}

	connectionErrorChannel := make(chan string)

	go elevio.PollObstructionSwitch(driverChannels.DrvObstr)
	go elevio.PollButtons(driverChannels.DrvButtons)
	go elevio.PollFloorSensor(driverChannels.DrvFloors)
	go elevio.PollStopButton(driverChannels.DrvStop)
	go FSM.Fsm(elevChannels, &elevator, driverChannels)

	go ordermanager.OrderMan(orderChannels, elevChannels, id, &elevator, connectionErrorChannel, &activeElevators, &elevatorArray)

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	//elevInt, _ := strconv.Atoi(elevPort)
	receiveInt, _ := strconv.Atoi(receivePort)
	transmitInt, _ := strconv.Atoi(transmitPort)
	receiveInt2, _ := strconv.Atoi(receivePort2)

	go peers.Transmitter(transmitInt+1, id, peerTxEnable)
	go peers.Receiver(receiveInt+1, peerUpdateCh)
	go peers.Receiver(receiveInt2+1, peerUpdateCh)

	networkTx := make(chan config.NetworkMessage)
	networkRx := make(chan config.NetworkMessage)

	//broadcasting and receiving to/from the first elevator
	go bcast.Transmitter(transmitInt, networkTx)
	go bcast.Receiver(receiveInt, networkRx)

	//broadcasting and receiving to/from the second elevator
	go bcast.Receiver(receiveInt2, networkRx)

	//Handles parsing and handling of messages sent and received
	go elevNet.SendElev(networkTx, elevChannels, id, orderChannels, &elevator)
	go elevNet.ReceiveElev(networkRx, elevChannels, peerUpdateCh, id, orderChannels, connectionErrorChannel, &activeElevators, &elevatorArray)

	//less go!!!!!
	FSM.InternalControl(driverChannels, orderChannels, elevChannels, &elevator)

}
