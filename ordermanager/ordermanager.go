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

func costFunc(elevatorArray [3]config.Elev, orderFloor int, activeElevators *[3]bool) string { //TODO: some less basic cost function maybe?, works OK though.
	closestDist := 1000.0 //just something large
	bestElevID := " "
	for elevIndx := 0; elevIndx < numElev; elevIndx++ {
		elev := elevatorArray[elevIndx]
		if math.Abs(float64(elev.Floor-orderFloor)) < float64(closestDist) && (elev.State != config.ERROR) && activeElevators[elevIndx] && elev.Floor >= 0 { //if in error state, do not receive more orders
			closestDist = math.Abs(float64(elev.Floor - orderFloor))
			bestElevID = strconv.Itoa(elevIndx + 1)
		}
	}
	println("closestDist: ", closestDist, "elev: ", bestElevID)
	return bestElevID

}

func transferHallOrders(lostElevator config.Elev, activeElevators *[3]bool, orderChan config.OrderChannels, id string,
	elevatorArray *[3]config.Elev, lostElevatorID string) {
	//delete(orderMap, lostElevatorID) //just to make sure
	println("In transfer hall order")
	for floor := 0; floor < numFloors; floor++ {
		for button := elevio.BT_HallUp; button < elevio.BT_Cab; button++ {
			if lostElevator.Queue[floor][button] {
				receivingID := costFunc(*elevatorArray, floor, activeElevators)
				order := elevio.ButtonEvent{Floor: floor, Button: button}
				receivingIdAsInt, _ := strconv.Atoi(receivingID)
				println("order transfered to: ", receivingID)

				if receivingID != id && activeElevators[receivingIdAsInt-1] { //only send order to an active elevator
					orderChan.SendOrder <- order
					orderChan.ExternalID <- receivingID
				} else { //if the chosen elevator is this one, send it directly
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
	activeElevators *[3]bool, elevatorArray *[3]config.Elev) {

	//orderMap := make(map[string]config.Elev) //map which keeps track of all elevators
	//orderMap[id] = *elev                     //insert this elevator into map with corresponding ID
	idAsInt, _ := strconv.Atoi(id)
	elevatorArray[idAsInt-1] = *elev
	for {
		select {
		case incomingOrder := <-orderChan.DelegateOrder:

			if incomingOrder.Button == elevio.BT_Cab { //cab orders are handled by the ordered elevator, always
				orderChan.ExtOrder <- incomingOrder
			} else {
				orderFloor := incomingOrder.Floor
				bestElevID := costFunc(*elevatorArray, orderFloor, activeElevators)

				if bestElevID == id { //if the chosen best elevator is this one, just send it to FSM
					orderChan.ExtOrder <- incomingOrder

				} else if bestElevID == " " {
					println("All elevators in error state. Hope they restart:)")
					orderChan.ExtOrder <- incomingOrder //send it to this elevator, because all elevators are equally bad to send to.
				} else { //if its one of the others, send it over the net
					orderChan.SendOrder <- incomingOrder
					orderChan.ExternalID <- bestElevID

				}
			}
			//fmt.Println("selected elev: ", bestElevID)
			/*
				case incMap := <-elevChan.MapChan: //update map
					for incId, incElev := range incMap { //TODO: maps give have floor == 0 when elevator starts between floors, needs fixing
						orderMap[incId] = incElev
					}*/

		case lostElevatorID := <-orderChan.LostConnection:
			println("lost evel ID: ", lostElevatorID)
			idAsInt, _ := strconv.Atoi(lostElevatorID)
			lostElevator := elevatorArray[idAsInt-1] //used to be map here
			go transferHallOrders(lostElevator, activeElevators, orderChan, id, elevatorArray, lostElevatorID)
		}
	}

}
