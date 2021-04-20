package elevNet

import (
	"fmt"
	"strconv"
	"time"

	"../config"
	"../driver/elevio"
	"../network/peers"
)

const sendingInterval = 100 * time.Millisecond //How often we send the current elevator object to the other elevators
const numOrders = 5                            //how many times we send each order to be robust against packet loss

//This function handles the sending of both the elevator object and the orders which will be sent over UDP
func SendElev(networkTx chan config.NetworkMessage, elevChan config.ElevChannels, id string, orderChan config.OrderChannels, elevator *config.Elev) {
	elev := config.Elev{}
	dummyButton := elevio.ButtonEvent{-1, elevio.BT_HallDown} //button to send when no order is included in message, or to set lights, but not take order
	turnOffLight := false
	sendingQueue := make(map[string]elevio.ButtonEvent) //in case the orders pile up, this map keeps all of them and sends them out in order
	for {
		select {
		case <-time.After(sendingInterval): //sends this elevator over UDP every *sendingInterval*
			timesSent := 0
			updateMessage := config.NetworkMessage{elev, id, false, dummyButton, false, turnOffLight}

			dummyButton = elevio.ButtonEvent{-1, elevio.BT_HallDown} //reset button and light bool
			turnOffLight = false
			for timesSent < 3 { //to be robust against packet loss. Set this to 1 if it doesnt matter
				networkTx <- updateMessage
				timesSent++
			}

		case elev = <-elevChan.Elevator:
			//update elevator

		case sendOrder := <-orderChan.SendOrder: //channel that has included an order to be sent externally
			timesSent := 0
			recipientID := <-orderChan.ExternalID
			sendingQueue[recipientID] = sendOrder
			for sendID, order := range sendingQueue {
				for timesSent < numOrders { //we send each message *numOrders* times to be robust against packetloss
					orderMessage := config.NetworkMessage{elev, sendID, true, order, true, false}
					networkTx <- orderMessage
					timesSent++
				}
				println("Sent ", sendID)
				delete(sendingQueue, sendID)
			}

		case completedOrder := <-orderChan.CompletedOrder:
			dummyButton = completedOrder //makes the next elevator object message include an order which is complete
			turnOffLight = true          //since order is complete, turn off the light at the order
		}

	}
}

//Function that handles the receiving of messages from the other elevators
func ReceiveElev(networkRx chan config.NetworkMessage, elevChan config.ElevChannels,
	peerUpdateCh chan peers.PeerUpdate, id string, orderChan config.OrderChannels,
	activeElevators *[config.NumElevs]bool, elevatorArray *[config.NumElevs]config.Elev) {
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			peerId, _ := strconv.Atoi(p.New)
			if peerId > 0 {
				activeElevators[peerId-1] = true
			} else {
				println("Connection broken!") //this happens when a connection is broken
			}

			if len(p.Lost) > 0 {
				for _, peer := range p.Lost {
					peerId, _ := strconv.Atoi(peer)
					if peer != id {
						activeElevators[peerId-1] = false //if a connection is lost, update the activeElevators
						orderChan.LostConnection <- peer  //sends a message which enables the lost elevatororders to be re-distributed
					}
				}
			}

		case receivedElev := <-networkRx:
			if receivedElev.ID == id && receivedElev.TakeOrder { //if an order to this ID is included, take the order
				orderChan.ExtOrder <- receivedElev.Order
			}
			if receivedElev.SetOrderLight { //if the message includes an order which another elevator has taken, set the order light
				elevio.SetButtonLamp(receivedElev.Order.Button, receivedElev.Order.Floor, true)
			}
			if receivedElev.TurnOffOrderLight { //if the message includes an order which another elevator has completed, turn off the order light
				elevio.SetButtonLamp(receivedElev.Order.Button, receivedElev.Order.Floor, false)
			}
			idAsInt, _ := strconv.Atoi(receivedElev.ID)
			elevatorArray[idAsInt-1] = receivedElev.Elevator

		case thisElev := <-elevChan.Elevator: //update the elevatorArray with this elevator
			idAsInt, _ := strconv.Atoi(id)
			elevatorArray[idAsInt-1] = thisElev
		}
	}
}
