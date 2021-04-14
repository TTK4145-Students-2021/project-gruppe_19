package ordermanager

import (
	"fmt"

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

func OrderMan(orderChan config.OrderChannels, elevChan config.ElevChannels) {

	for {
		select {
		case incomingOrder := <-orderChan.DelegateOrder:
			//othersLocation := <-orderChan.OthersLocation
			//selectedElev := costFunc(incomingOrder, othersLocation)
			fmt.Println("selected elev: ", 1)
			orderChan.ExtOrder <- incomingOrder

		case elevState := <-elevChan.Elevator:
			if iteration%10000000 == 0 {
				println("in order manager", elevState.Floor)
			}

			iteration++
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
