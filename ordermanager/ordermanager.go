package ordermanager

import (
	"math"
	"strconv"

	"../config"
	"../driver/elevio"
)

const numElev = 3
const numButtons = 3
const numFloors = 4

func costFunc(orderMap map[string]config.Elev, orderFloor int) string { //TODO: some less basic cost function maybe?, works OK though.
	closestDist := 1000.0 //just something large
	bestElevID := " "
	for id, elev := range orderMap {
		if math.Abs(float64(elev.Floor-orderFloor)) < float64(closestDist) && (elev.State != config.ERROR) { //if in error state, do not receive more orders
			closestDist = math.Abs(float64(elev.Floor - orderFloor))
			bestElevID = id
		}
	}
	println("closestDist: ", closestDist)
	return bestElevID

}

func transferHallOrders(lostElevator config.Elev, activeElevators *[3]bool, orderChan config.OrderChannels, id string, orderMap map[string]config.Elev) {
	for floor := 0; floor < numFloors; floor++ {
		for button := elevio.BT_HallUp; button < elevio.BT_Cab; button++ {
			if lostElevator.Queue[floor][button] {
				receivingID := costFunc(orderMap, floor)
				order := elevio.ButtonEvent{Floor: floor, Button: button}
				receivingIdAsInt, _ := strconv.Atoi(receivingID)
				if receivingID != id && activeElevators[receivingIdAsInt-1] { //this assumes map and active elevators agree
					orderChan.SendOrder <- order
					orderChan.ExternalID <- receivingID
				} else {
					orderChan.ExtOrder <- order
					//elevatorList[lostElevatorIndex-1].Queue[floor][button] = false
				}

			}
		}
	}
}

func printMap(orderMap map[string]config.Elev) {
	for id, elev := range orderMap {
		println("ElevatorID: ", id)
		println("Elevator Floor: ", elev.Floor)
	}
}

func OrderMan(orderChan config.OrderChannels, elevChan config.ElevChannels, id string, elev *config.Elev, connErrorChan chan string,
	activeElevators *[3]bool) {

	orderMap := make(map[string]config.Elev) //map which keeps track of all elevators
	orderMap[id] = *elev                     //insert this elevator into map with corresponding ID
	for {
		select {
		case incomingOrder := <-orderChan.DelegateOrder:

			if incomingOrder.Button == elevio.BT_Cab { //cab orders are handled by the ordered elevator, always
				orderChan.ExtOrder <- incomingOrder
			}
			printMap(orderMap)
			orderFloor := incomingOrder.Floor
			bestElevID := costFunc(orderMap, orderFloor)

			if bestElevID == id { //if the chosen best elevator is this one, just send it to FSM
				orderChan.ExtOrder <- incomingOrder

			} else if bestElevID == " " {
				println("All elevators in error state. Hope they restart:)")
				orderChan.ExtOrder <- incomingOrder //send it to this elevator, because all elevators are equally bad to send to.
			} else { //if its one of the others, send it over the net
				orderChan.SendOrder <- incomingOrder
				orderChan.ExternalID <- bestElevID

			}
			//fmt.Println("selected elev: ", bestElevID)

		case incMap := <-elevChan.MapChan: //update map
			for incId, incElev := range incMap { //TODO: maps give have floor == 0 when elevator starts between floors, needs fixing
				orderMap[incId] = incElev
			}

		case lostElevatorID := <-orderChan.LostConnection:
			lostElevator := orderMap[lostElevatorID]
			delete(orderMap, lostElevatorID)
			go transferHallOrders(lostElevator, activeElevators, orderChan, id, orderMap)

			/*

				case errorID := <-connErrorChan:
					errorElev := orderMap[errorID]
					delete(orderMap, errorID)
					println("elev deleted")
					printMap(orderMap)
					ordersLost := make([]elevio.ButtonEvent, numButtons*numFloors)
					orderIndex := 0
					for btn := 0; btn < numButtons; btn++ {
						for floor := 0; floor < numFloors; floor++ {
							if errorElev.Queue[floor][btn] {
								order := elevio.ButtonEvent{
									Floor:  floor,
									Button: elevio.IntToButtonType(btn),
								}
								ordersLost[orderIndex] = order
								orderIndex++
							}
						}
					}
					receivingID := " "
					for ID, _ := range orderMap {
						receivingID = ID //just takes the last one, doesnt matter which one it is
					}
					if receivingID == id {
						for order := range ordersLost {
							orderChan.ExtOrder <- ordersLost[order]
						}
					} else {
						for order := range ordersLost {
							orderChan.SendOrder <- ordersLost[order]
							orderChan.ExternalID <- receivingID
						}
					}

					//remove from oderMap
					//delegate orders from this elevator
			*/
		}
	}

}
