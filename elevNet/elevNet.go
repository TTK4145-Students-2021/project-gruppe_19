package elevNet

import (
	"fmt"
	"strconv"
	"time"

	"../config"
	"../driver/elevio"
	"../network/peers"
)

const numFloors = 4
const numButtons = 3

const interval = 100 * time.Millisecond //should sync with elevator channel timing

const numElevs = 3

var iteration = 0

func SendElev(networkTx chan config.NetworkMessage, elevChan config.ElevChannels, id string, orderChan config.OrderChannels, elevator *config.Elev) {
	elev := config.Elev{}
	dummyButton := elevio.ButtonEvent{-1, elevio.BT_HallDown} //button to send when no order is included in message
	for {
		select {
		case <-time.After(interval): //sends this elevator over UDP every *interval*
			updateMessage := config.NetworkMessage{elev, id, false, dummyButton, false}
			networkTx <- updateMessage

		case elev = <-elevChan.Elevator:
			//update elevator

		case sendOrder := <-orderChan.SendOrder: //channel that has included an order to be sent externally
			recipientID := <-orderChan.ExternalID                                            //not sure if this works. might have sync issues. update: think it works :)
			orderMessage := config.NetworkMessage{elev, recipientID, true, sendOrder, false} //sends current elevator with an order
			networkTx <- orderMessage
		}

	}
}

func ReceiveElev(networkRx chan config.NetworkMessage, elevChan config.ElevChannels,
	peerUpdateCh chan peers.PeerUpdate, id string, orderChan config.OrderChannels, connErrorChan chan string,
	activeElevators *[3]bool, elevatorArray *[3]config.Elev) {
	//elevMap := make(map[string]config.Elev)

	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			//updates activeElevs with new connection
			peerId, _ := strconv.Atoi(p.New)
			if peerId > 0 {
				activeElevators[peerId-1] = true
			} else {
				println("Connection broken!") //this happens when a connection is broken
			}
			//If lost a peer, update the active elevator list and start order redistribution
			if len(p.Lost) > 0 {
				for _, peer := range p.Lost {
					peerId, _ := strconv.Atoi(peer)
					if peer != id {
						activeElevators[peerId-1] = false
						orderChan.LostConnection <- peer
					}
				}
			}

		case receivedElev := <-networkRx:
			if receivedElev.ID == id && receivedElev.OrderIncl {
				orderChan.ExtOrder <- receivedElev.Order
			}
			idAsInt, _ := strconv.Atoi(receivedElev.ID)
			//println("Received elevator ID, FLOOR: ", receivedElev.ID, " ", receivedElev.Elevator.Floor)
			elevatorArray[idAsInt-1] = receivedElev.Elevator

			//elevMap[receivedElev.ID] = receivedElev.Elevator //update elevatorMap
			//elevChan.MapChan <- elevMap

		case thisElev := <-elevChan.Elevator:
			idAsInt, _ := strconv.Atoi(id)
			elevatorArray[idAsInt-1] = thisElev

			if iteration%20 == 0 {
				for i := 0; i < 3; i++ {
					println("Elev ", i+1, " floor: ", elevatorArray[i].Floor)
				}
				println(" ")
			}
			iteration++
			//map update here before
		}
	}
}
