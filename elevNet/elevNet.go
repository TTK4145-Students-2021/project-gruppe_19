package elevNet

import (
	"fmt"
	"time"

	"../config"
	"../network/peers"
)

func SendElev(networkTx chan config.NetworkMessage, elevChan config.ElevChannels) {
	for {
		elev := <-elevChan.Elevator
		netMessage := config.NetworkMessage{elev}
		networkTx <- netMessage
		println("transmitting")
		time.Sleep(1 * time.Second)
	}
}

func ReceiveElev(networkRx chan config.NetworkMessage, elevChan config.ElevChannels, peerUpdateCh chan peers.PeerUpdate) {
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
