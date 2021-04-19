package ordermanager

import (
	"strconv"

	// "p/FSM"
	"p/FSM"
	"p/config"
	"p/driver/elevio"
)

// func costFunc(elevatorArray [config.NumElevs]config.Elev, activeElevators *[config.NumElevs]bool, order elevio.ButtonEvent) string { //TODO: some less basic cost function maybe?, works OK though.
// 	orderFloor := order.Floor
// 	closestDist := 1000.0 //just something large
// 	bestElevID := " "
// 	for elevIndx := 0; elevIndx < config.NumElevs; elevIndx++ {
// 		elev := elevatorArray[elevIndx]
// 		if math.Abs(float64(elev.Floor-orderFloor)) < float64(closestDist) &&
// 			(elev.State != config.ERROR) && activeElevators[elevIndx] && elev.Floor >= 0 { //if in error state, do not receive more orders
// 			closestDist = math.Abs(float64(elev.Floor - orderFloor))
// 			bestElevID = strconv.Itoa(elevIndx + 1)
// 		}
// 	}
// 	println("closestDist: ", closestDist, "elev: ", bestElevID)
// 	return bestElevID

// }

func atOrderFloor(elevator config.Elev, order elevio.ButtonEvent) bool {
	println(elevator.Dir)
	// println(order.Button)
	if elevator.Dir == config.UP && order.Button == elevio.BT_HallUp {
		if elevator.Floor == order.Floor {
			// println("true")
			return true
		}
	} else if elevator.Dir == config.DOWN && order.Button == elevio.BT_HallDown {
		if elevator.Floor == order.Floor {
			// println("true")
			return true
		}
	} else if elevator.Dir == config.STILL {
		return true
	} else if elevator.Dir == config.DOWN && (order.Button == 0) && !FSM.OrdersBelow(elevator) {
		if elevator.Floor == order.Floor {
			// println("true")
			return true
		}
	} else if elevator.Dir == config.UP && (order.Button == 1) && !FSM.OrdersAbove(elevator) {
		if elevator.Floor == order.Floor {
			// println("true")
			return true
		}
	}
	return false
}

func chooseDirection(elevator config.Elev) config.Direction {
	// println("floor:  ", elevator.Floor, "  Dir:   ", elevator.Dir)
	switch elevator.Dir {
	case config.UP:
		if FSM.OrdersAbove(elevator) {
			return config.UP
		} else if FSM.OrdersBelow(elevator) {
			return config.DOWN
		} else {
			return config.STILL
		}
	case config.DOWN:
		if FSM.OrdersBelow(elevator) {
			println("going down")
			return config.DOWN
		} else if FSM.OrdersAbove(elevator) {
			return config.UP
		} else {
			return config.STILL
		}
	case config.STILL:
		if FSM.OrdersAbove(elevator) {
			return config.UP
		} else if FSM.OrdersBelow(elevator) {
			return config.DOWN
		} else {
			return config.STILL
		}
	}
	return config.STILL
}

func dirToButton(dir config.Direction) elevio.ButtonType {
	switch dir {
	case config.UP:
		return elevio.BT_HallUp
	case config.DOWN:
		return elevio.BT_HallDown
	case config.STILL:
	}
	return 2
}

func timeToCompleteOrder(elevator config.Elev, order elevio.ButtonEvent, id int) int {
	var elev config.Elev
	elev = elevator
	if elev.Floor < 0 {
		elev.Floor = 0
	}
	duration := 0
	elev.Queue[order.Floor][order.Button] = true

	switch elevator.State {
	case config.IDLE:
		elev.Dir = chooseDirection(elev)
		if elev.Dir == config.STILL {
			return duration
		}
		break
	case config.RUNNING:
		duration += config.TimeToMove / 2
		elev.Floor += int(elev.Dir)
		break
	case config.DOOR_OPEN:
		duration -= config.TimeToOpenDoors
	}
	elev.Dir = chooseDirection(elev)
	println("elevator:   ", id, "floor:  ", elev.Floor, "  Dir:   ", elev.Dir, "orderfloor:   ", order.Floor)
	for {
		// println("elevator:   ", id, "floor:  ", elevator.Floor, "  Dir:   ", elevator.Dir)
		if FSM.OrdersInFloor(elev) {
			println("in first if")

			if atOrderFloor(elev, order) {
				println("in second if")
				return duration
			}
			elev.Queue[elev.Floor][dirToButton((elev.Dir))] = false
			duration += config.TimeToOpenDoors
			elev.Dir = chooseDirection(elev)
		}
		elev.Queue[order.Floor][order.Button] = true
		elev.Floor += int(elev.Dir)
		println("elevator:   ", id, "floor:  ", elev.Floor, "  Dir:   ", elev.Dir, "orderfloor:   ", order.Floor)
		// println(elev.Dir)
		duration += config.TimeToMove
		// elev.Dir = chooseDirection(elev)
	}
}

func costFunc(elevatorArray [config.NumElevs]config.Elev, activeElevators *[config.NumElevs]bool, order elevio.ButtonEvent) string {
	bestTimeToComplete := 1000
	var bestId string

	for elevatorIndex := 0; elevatorIndex < config.NumElevs; elevatorIndex++ {
		// println(elevatorIndex, "  ", activeElevators[elevatorIndex])
		if activeElevators[elevatorIndex] {
			timeToComplete := timeToCompleteOrder(elevatorArray[elevatorIndex], order, elevatorIndex)
			println("done ttc")
			if timeToComplete < bestTimeToComplete {
				bestTimeToComplete = timeToComplete
				bestId = strconv.Itoa(elevatorIndex + 1)
			}
		}
	}
	println("closestDist: ", bestTimeToComplete, "elev: ", bestId)
	return bestId
}

func transferOrders(lostElevator config.Elev, activeElevators *[config.NumElevs]bool, orderChan config.OrderChannels, id string,
	elevatorArray *[config.NumElevs]config.Elev, lostElevatorID string) {
	var order elevio.ButtonEvent
	for floor := 0; floor < config.NumFloors; floor++ {
		for button := elevio.BT_HallUp; button < elevio.BT_Cab; button++ {
			if lostElevator.Queue[floor][button] {
				order.Button = button
				order.Floor = floor
				receivingID := costFunc(*elevatorArray, activeElevators, order)
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

func OrderMan(orderChan config.OrderChannels, elevChan config.ElevChannels, id string, elev *config.Elev, connErrorChan chan string,
	activeElevators *[config.NumElevs]bool, elevatorArray *[config.NumElevs]config.Elev) {
	idAsInt, _ := strconv.Atoi(id)
	elevatorArray[idAsInt-1] = *elev
	for {
		select {
		case incomingOrder := <-orderChan.DelegateOrder:

			if incomingOrder.Button == elevio.BT_Cab { //cab orders are handled by the ordered elevator, always
				orderChan.ExtOrder <- incomingOrder
			} else {
				// orderFloor := incomingOrder.Floor
				bestElevID := costFunc(*elevatorArray, activeElevators, incomingOrder)

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

		case lostElevatorID := <-orderChan.LostConnection:
			println("lost elev ID: ", lostElevatorID)
			idAsInt, _ := strconv.Atoi(lostElevatorID)
			lostElevator := elevatorArray[idAsInt-1]
			go transferOrders(lostElevator, activeElevators, orderChan, id, elevatorArray, lostElevatorID)
		}
	}

}
