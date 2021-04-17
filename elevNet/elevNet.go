package elevNet

import (
	"fmt"
	"time"

	"../config"
	"../driver/elevio"
	"../network/peers"
)

const numFloors = 4
const numButtons = 3

const interval = 1000 * time.Millisecond

//const timeout = 500 * time.Millisecond

func SendElev(networkTx chan config.NetworkMessage, elevChan config.ElevChannels, id string, orderChan config.OrderChannels, elevator *config.Elev) {
	elev := *elevator
	dummyButton := elevio.ButtonEvent{-1, elevio.BT_HallDown} //button to send when no order is included in message
	for {
		select {
		case <-time.After(interval): //sends this elevator over UDP every *interval*
			updateMessage := config.NetworkMessage{elev, id, false, dummyButton}
			networkTx <- updateMessage

		case elev = <-elevChan.Elevator:
			//update elevator

		case sendOrder := <-orderChan.SendOrder: //channel that has included an order to be sent externally
			recipientID := <-orderChan.ExternalID                                     //not sure if this works. might have sync issues. update: think it works :)
			orderMessage := config.NetworkMessage{elev, recipientID, true, sendOrder} //sends current elevator with an order
			networkTx <- orderMessage
		}

	}
}

func ReceiveElev(networkRx chan config.NetworkMessage, elevChan config.ElevChannels,
	peerUpdateCh chan peers.PeerUpdate, id string, orderChan config.OrderChannels) {
	elevMap := make(map[string]config.Elev)
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case receivedElev := <-networkRx:
			if receivedElev.ID == id && receivedElev.OrderIncl {
				orderChan.ExtOrder <- receivedElev.Order
			}

			elevMap[receivedElev.ID] = receivedElev.Elevator
			elevChan.MapChan <- elevMap

		case thisElev := <-elevChan.Elevator:
			elevMap[id] = thisElev
			elevChan.MapChan <- elevMap
		}
	}
}
