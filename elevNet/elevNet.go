package elevNet

import (
	"fmt"
	"strconv"
	"time"

	"../config"
	"../driver/elevio"
	"../network/peers"
)

const sendingInterval = 100 * time.Millisecond //should sync with elevator channel timing?
const orderInterval = 10

func SendElev(networkTx chan config.NetworkMessage, elevChan config.ElevChannels, id string, orderChan config.OrderChannels, elevator *config.Elev) {
	elev := config.Elev{}
	dummyButton := elevio.ButtonEvent{-1, elevio.BT_HallDown} //button to send when no order is included in message, or to set lights, but not take order
	turnOffLight := false
	sendingQueue := make(map[string]elevio.ButtonEvent)
	for {
		select {
		case <-time.After(sendingInterval): //sends this elevator over UDP every *interval*
			updateMessage := config.NetworkMessage{elev, id, false, dummyButton, false, turnOffLight}

			dummyButton = elevio.ButtonEvent{-1, elevio.BT_HallDown} //reset button and light bool
			turnOffLight = false
			networkTx <- updateMessage

		case elev = <-elevChan.Elevator:
			//update elevator

		case sendOrder := <-orderChan.SendOrder: //channel that has included an order to be sent externally
			recipientID := <-orderChan.ExternalID
			sendingQueue[recipientID] = sendOrder
			for sendID, order := range sendingQueue {
				go func() { //this works. Not sure if intended, but elevators still run
					orderMessage := config.NetworkMessage{elev, sendID, true, order, true, false}
					networkTx <- orderMessage
					time.Sleep(orderInterval * time.Millisecond)
				}()
				println("Sent ", sendID)
				delete(sendingQueue, sendID)
			}

			//not sure if this works. might have sync issues. update: think it works :)
			orderMessage := config.NetworkMessage{elev, recipientID, true, sendOrder, true, false} //sends current elevator with an order
			networkTx <- orderMessage

		case completedOrder := <-orderChan.CompletedOrder:
			dummyButton = completedOrder
			turnOffLight = true
		}

	}
}

func ReceiveElev(networkRx chan config.NetworkMessage, elevChan config.ElevChannels,
	peerUpdateCh chan peers.PeerUpdate, id string, orderChan config.OrderChannels,
	activeElevators *[config.NumElevs]bool, elevatorArray *[config.NumElevs]config.Elev) {
	//elevMap := make(map[string]config.Elev)

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
						activeElevators[peerId-1] = false
						orderChan.LostConnection <- peer
					}
				}
			}

		case receivedElev := <-networkRx:
			if receivedElev.ID == id && receivedElev.TakeOrder {
				orderChan.ExtOrder <- receivedElev.Order
			}
			if receivedElev.SetOrderLight {
				elevio.SetButtonLamp(receivedElev.Order.Button, receivedElev.Order.Floor, true)
			}
			if receivedElev.TurnOffOrderLight {
				elevio.SetButtonLamp(elevio.BT_HallDown, receivedElev.Order.Floor, false)
				elevio.SetButtonLamp(elevio.BT_HallUp, receivedElev.Order.Floor, false)
			}
			idAsInt, _ := strconv.Atoi(receivedElev.ID)
			elevatorArray[idAsInt-1] = receivedElev.Elevator

		case thisElev := <-elevChan.Elevator:
			idAsInt, _ := strconv.Atoi(id)
			elevatorArray[idAsInt-1] = thisElev
		}
	}
}
