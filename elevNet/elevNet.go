package elevNet

import (
	"fmt"
	"time"

	"../config"
	"../network/peers"
)

func SendElev(networkTx chan config.NetworkMessage, elevChan config.ElevChannels, id string) {
	for {
		elev := <-elevChan.Elevator
		netMessage := config.NetworkMessage{elev, id}
		networkTx <- netMessage
		println("transmitting")
		time.Sleep(1 * time.Second)
	}
}

func ReceiveElev(networkRx chan config.NetworkMessage, elevChan config.ElevChannels,
	peerUpdateCh chan peers.PeerUpdate, id string, mapChan chan map[string]config.Elev) {
	elevMap := make(map[string]config.Elev)
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case receivedElev := <-networkRx:
			fmt.Print("received ID ", receivedElev.ID)
			elevMap[receivedElev.ID] = receivedElev.Elevator
			mapChan <- elevMap

		case thisElev := <-elevChan.Elevator:
			elevMap[id] = thisElev
			mapChan <- elevMap
		}
	}
}
