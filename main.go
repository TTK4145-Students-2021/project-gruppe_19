package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"./FSM"
	"./config"
	"./driver/elevio"
	"./network/bcast"
	"./network/localip"
	"./network/peers"
	"./ordermanager"
)

const numFloors = 4
const numButtons = 3
const numElevs = 3

type NetworkMessage struct {
	Elevator FSM.Elev
}
type Elev struct {
	State FSM.ElevatorState
	Dir   FSM.Direction
	Floor int
	Queue [numFloors][numButtons]bool
}

type ElevChannels struct{
	Elevator chan Elev
}


func main() {

	var elevator = FSM.Elev{
		State: FSM.IDLE,
		Dir:   FSM.STILL,
		Floor: 0, //denne har ingenting å si siden den oppdateres i FSMinit
		Queue: [numFloors][numButtons]bool{},
	}

	var id string
	elevPort_p := flag.String("elev_port", "15657", "The port of the elevator to connect to (for sim purposes)")
	//får noen ganger out of index error i ordersInFloor funksjonen når forskjellige porter brukes!!!??!

	flag.StringVar(&id, "id", "", "id of this peer")
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

	elevChannels := ElevChannels{
		Elevator: make(chan Elev)
	}





	/*dummyString := "fuck u bitch"
	transmitEnable := make(chan bool)
	go peers.Transmitter(22349, dummyString, transmitEnable)*/

	go elevio.PollObstructionSwitch(driverChannels.DrvObstr)
	go elevio.PollButtons(driverChannels.DrvButtons)
	go elevio.PollFloorSensor(driverChannels.DrvFloors)
	go elevio.PollStopButton(driverChannels.DrvStop)
	go FSM.Fsm(driverChannels.DoorsOpen, elevChannels)

	go ordermanager.OrderMan(orderChannels)

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
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	networkTx := make(chan NetworkMessage)
	networkRx := make(chan NetworkMessage)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	tranPort := strconv.Atoi(*elevPort_p)
	go bcast.Transmitter(tranPort+1, networkTx)
	go bcast.Receiver(tranPort+1, networkRx)

	elevChan := make(chan FSM.Elev)



	// The example message. We just send one of these every second.
	go func() {
		elevChan <- elevator
		netMessage := elevChan
		for {
			networkTx <- netMessage
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	go func(){
		for {
			select {
			case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			case receivedElev := <-networkRx:
			fmt.Print(receivedElev)
			}
		}
	}


	FSM.InternalControl(driverChannels, orderChannels, elevChannels)

}
