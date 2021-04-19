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

const interval = 1000 * time.Millisecond

const timerTime = 3

const numElevs = 3

var connectionTimer1 = time.NewTimer(timerTime * time.Second)
var connectionTimer2 = time.NewTimer(timerTime * time.Second)

var timerArray = [numElevs - 1]*time.Timer{connectionTimer1, connectionTimer2} //hardkoda, kan moduleres med slices
var timerShot = [numElevs - 1]bool{false, false}

func SendElev(networkTx chan config.NetworkMessage, elevChan config.ElevChannels, id string, orderChan config.OrderChannels, elevator *config.Elev) {
	elev := *elevator
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
	peerUpdateCh chan peers.PeerUpdate, id string, orderChan config.OrderChannels, connErrorChan chan string, activeElevators *[3]bool) {
	elevMap := make(map[string]config.Elev)
	//connectionErrorMap := make(map[string]*time.Timer)
	connectionErrorIndex := make(map[string]int)
	connIndex := 0       //not currently used
	timerArray[0].Stop() //redo
	timerArray[1].Stop()
	//idIndex, _ := strconv.Atoi(id)

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

			/*if elevatorList[idIndex-1].ID == id {
				ch.TransmittStateCh <- map[string][cf.NumElevators]cf.ElevatorState{idAsString: *elevatorList}
			}*/

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
			_, ok := connectionErrorIndex[receivedElev.ID] //checks if an index already exists for this elevator ID
			if !ok && receivedElev.ID != id {              //if it doesnt, make one
				connectionErrorIndex[receivedElev.ID] = connIndex
				connIndex++
			}

			timerIndex := connectionErrorIndex[receivedElev.ID]   //index in timerArray for the received elevator
			timerArray[timerIndex].Reset(timerTime * time.Second) // reset connection timer for the received elevator
			elevMap[receivedElev.ID] = receivedElev.Elevator      //update elevatorMap
			elevChan.MapChan <- elevMap

		case thisElev := <-elevChan.Elevator:
			elevMap[id] = thisElev
			elevChan.MapChan <- elevMap
			for i := 0; i < 3; i++ {
				println("Elev: ", i+1, " ", activeElevators[i])
			}
			/*
				case <-timerArray[0].C: //chooses a timer to check each iteration

					if !timerShot[0] {
						thisIndex := 0 //randIndexCounter % (connIndex)
						errorID := " "
						for elevID, indx := range connectionErrorIndex {
							if thisIndex == indx {
								errorID = elevID
							}
						}
						if errorID != " " { //just in case something weird happens
							println("lost connection to: ", errorID)
							connErrorChan <- errorID
						} else {
							println("This should never happen. Connection ErrorID is undefined")
						}
						timerShot[0] = true
					}

				case <-timerArray[1].C:
					if !timerShot[1] {

						thisIndex := 1 //randIndexCounter % (connIndex)
						errorID := " "
						for elevID, indx := range connectionErrorIndex {
							if thisIndex == indx {
								errorID = elevID
							}
						}
						if errorID != " " { //just in case something weird happens
							println("lost connection to: ", errorID)
							connErrorChan <- errorID
						} else {
							println("This should never happen. Connection ErrorID is undefined")
						}
						timerShot[1] = true
					}*/
		}

	}
}
