package elevNet

import (
	"fmt"

	"../config"
	"../driver/elevio"
	"../network/peers"
)

const numFloors = 4
const numButtons = 3

func SendElev(networkTx chan config.NetworkMessage, elevChan config.ElevChannels, id string, orderChan config.OrderChannels) {
	elev := config.Elev{config.IDLE, config.UP, 0, [numFloors][numButtons]bool{}}
	for {
		dummyButton := elevio.ButtonEvent{-1, elevio.BT_HallDown}
		netMessage := config.NetworkMessage{elev, id, false, dummyButton}
		networkTx <- netMessage
		select {
		case elev = <-elevChan.Elevator:

		case sendOrder := <-orderChan.SendOrder:
			recipientID := " "
			if id == "forste" {
				recipientID = "andre"
			} else {
				recipientID = "forste"
			}
			//need logic to send to correct elevator, this only works with two
			netMessage := config.NetworkMessage{elev, recipientID, true, sendOrder} //bloat? idk
			networkTx <- netMessage
		}

	}
}

func ReceiveElev(networkRx chan config.NetworkMessage, elevChan config.ElevChannels,
	peerUpdateCh chan peers.PeerUpdate, id string, mapChan chan map[string]config.Elev, orderChan config.OrderChannels) {
	elevMap := make(map[string]config.Elev)
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case receivedElev := <-networkRx:
			if receivedElev.OrderIncl { //if message includes order, send it to FSM
				orderChan.ExtOrder <- receivedElev.Order
			}

			elevMap[receivedElev.ID] = receivedElev.Elevator
			mapChan <- elevMap

		case thisElev := <-elevChan.Elevator:
			println("received elevator channel")
			elevMap[id] = thisElev
			mapChan <- elevMap
		}
	}
}
