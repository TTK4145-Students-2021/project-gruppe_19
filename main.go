package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"./FSM"
	"./config"
	"./driver/elevio"
	"./elevNet"
	"./network/bcast"
	"./network/localip"
	"./network/peers"
	"./ordermanager"
)

func main() {

	const numFloors = 4
	const numButtons = 3

	var elevator = config.Elev{
		State: config.IDLE,
		Dir:   config.STILL,
		Floor: 0, //denne har ingenting å si siden den oppdateres i FSMinit
		Queue: [numFloors][numButtons]bool{},
	}

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
		ExtOrder:      make(chan elevio.ButtonEvent),
		DelegateOrder: make(chan elevio.ButtonEvent),
		SendOrder:     make(chan elevio.ButtonEvent),
		ExternalID:    make(chan string),
	}

	elevChannels := config.ElevChannels{
		Elevator: make(chan config.Elev),
	}

	mapChan := make(chan map[string]config.Elev)

	go elevio.PollObstructionSwitch(driverChannels.DrvObstr)
	go elevio.PollButtons(driverChannels.DrvButtons)
	go elevio.PollFloorSensor(driverChannels.DrvFloors)
	go elevio.PollStopButton(driverChannels.DrvStop)
	go FSM.Fsm(driverChannels.DoorsOpen, elevChannels, &elevator)

	go ordermanager.OrderMan(orderChannels, elevChannels, mapChan, id, &elevator)

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	//elevInt, _ := strconv.Atoi(elevPort)
	receiveInt, _ := strconv.Atoi(receivePort)
	transmitInt, _ := strconv.Atoi(transmitPort)
	receiveInt2, _ := strconv.Atoi(receivePort2)

	go peers.Transmitter(transmitInt+1, id, peerTxEnable)
	go peers.Receiver(receiveInt+1, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	networkTx := make(chan config.NetworkMessage)
	networkRx := make(chan config.NetworkMessage)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.

	//broadcasting and receiving to/from the first elevator
	go bcast.Transmitter(transmitInt, networkTx)
	go bcast.Receiver(receiveInt, networkRx)

	//broadcasting and receiving to/from the second elevator
	go bcast.Transmitter(transmitInt, networkTx)
	go bcast.Receiver(receiveInt2, networkRx)

	//Handles parsing and handling of messages sent and received
	go elevNet.SendElev(networkTx, elevChannels, id, orderChannels)
	go elevNet.ReceiveElev(networkRx, elevChannels, peerUpdateCh, id, mapChan, orderChannels)

	//less go
	FSM.InternalControl(driverChannels, orderChannels, elevChannels, &elevator)

}
