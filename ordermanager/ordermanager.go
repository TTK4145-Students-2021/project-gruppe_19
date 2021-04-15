package ordermanager

import (
	"fmt"
	"math"

	"../config"
	"../driver/elevio"
)

const numElev = 3

/*
type orderMess struct {
	Floor     int
	Direction int
	Timelist  []int
}

//Wait for new hall orders
func ordermanager() {

	drv_buttons := make(chan elevio.ButtonEvent)
	orderMessage := make(chan orderMess)
	var hallbtn elevio.ButtonEvent
	var order orderMess

	for {
		select {
		case hallbtn = <-drv_buttons:
			if hallbtn.Button == 2 {
			} else {

			}
		case order = <-orderMessage:
		}
	}
}*/

func costFunc(incomingOrder elevio.ButtonEvent, othersLocation [numElev]int) int {
	fmt.Println("inside ", incomingOrder, othersLocation)
	return 1
}

var iteration = 0

func OrderMan(orderChan config.OrderChannels, elevChan config.ElevChannels, mapChan chan map[string]config.Elev) {

	for {
		select {
		case incomingOrder := <-orderChan.DelegateOrder:
			//othersLocation := <-orderChan.OthersLocation
			//selectedElev := costFunc(incomingOrder, othersLocation)
			fmt.Println("herja")
			elevMap := <-mapChan

			orderFloor := incomingOrder.Floor
			closestDist := 1000.0
			bestElevID := " "

			for id, elev := range elevMap {
				if math.Abs(float64(elev.Floor-orderFloor)) < float64(closestDist) {
					closestDist = math.Abs(float64(elev.Floor - orderFloor))
					bestElevID = id
				}
			}

			fmt.Println("selected elev: ", bestElevID)
			orderChan.ExtOrder <- incomingOrder
		case elevState := <-elevChan.Elevator: //something needs to take in the channels all the time, or else the FSM gets stuck
			<-elevChan.Elevator
			if iteration%10000000 == 0 {
				println("in order manager", elevState.Floor)
			}

			iteration++ //dette er bare piss for å ta inn en elevator hele tiden. Skal fjernes

		}
	}

}

//Start groutine to handle order

//Do cost calculation (send to cost function)

//If itself: send to fsm and broadcast
//Else: send to correct elevator, and monitor.

//Wait for order broadcast

//Start goroutine to deal with broadcast

//If taking order: send to fsm (wait 1 sec) and broadcast order accept
//Else: Wait for order accept, then wait for order finish
