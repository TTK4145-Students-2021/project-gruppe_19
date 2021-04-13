package FSM

import (
	"fmt"

	"../config"
	"../driver/elevio"
)

type Keypress struct {
	Floor              int
	Btn                elevio.ButtonType
	DesignatedElevator int
	Done               bool
}

func ordersAbove(elevator config.Elev) bool {
	currentFloor := elevator.Floor
	for i := currentFloor + 1; i < numFloors; i++ {
		if elevator.Queue[i][0] || elevator.Queue[i][1] || elevator.Queue[i][2] {
			return true
		}
	}
	return false
}

func ordersBelow(elevator config.Elev) bool {
	currentFloor := elevator.Floor
	for i := currentFloor - 1; i > -1; i-- {
		if elevator.Queue[i][0] || elevator.Queue[i][1] || elevator.Queue[i][2] {
			return true
		}
	}
	return false
}

func ordersInFloor(elevator config.Elev) bool {
	for btn := 0; btn < numButtons; btn++ {
		if elevator.Queue[elevator.Floor][btn] {
			if elevator.Dir == config.UP && btn == 0 { //makes sure the elevator only stops of the order is in the same direction
				return true
			} else if elevator.Dir == config.DOWN && btn == 1 { //makes sure the elevator only stops of the order is in the same direction
				return true
			} else if btn == 2 { //cab orders will always stop no matter which direction
				return true
			} else if elevator.Floor == 0 || elevator.Floor == (numFloors-1) { //takes care of the edge cases
				return true
			} else {
				return false
			}

		}
	}
	return false
}

func DeleteOrder(elevator *config.Elev) {
	deletedFloor := elevator.Floor
	for i := 0; i < numButtons; i++ {
		elevator.Queue[elevator.Floor][i] = false
	}
	fmt.Println("Order deleted at ", deletedFloor)
}

func DeleteAllOrders(elevator *config.Elev) {
	for btn := 0; btn < numButtons; btn++ {
		for floor := 0; floor < numFloors; floor++ {
			elevator.Queue[floor][btn] = false
			fmt.Println(elevator.Queue[floor][btn])
		}
	}

}

func chooseElevatorDir(elevator config.Elev) elevio.MotorDirection {
	switch elevator.Dir {
	case config.STILL:
		if ordersAbove(elevator) {
			return elevio.MD_Up
		} else if ordersBelow(elevator) {
			return elevio.MD_Down
		} else {
			return elevio.MD_Stop
		}
	case config.UP:
		if ordersAbove(elevator) {
			return elevio.MD_Up
		} else if ordersBelow(elevator) {
			return elevio.MD_Down
		} else {
			return elevio.MD_Stop
		}

	case config.DOWN:
		if ordersBelow(elevator) {
			return elevio.MD_Down
		} else if ordersAbove(elevator) {
			return elevio.MD_Up
		} else {
			return elevio.MD_Stop
		}
	}
	return elevio.MD_Stop
}

func shouldStop(elevator config.Elev) bool {
	switch elevator.Dir {
	case config.UP:
		return elevator.Queue[elevator.Floor][elevio.BT_HallUp] ||
			elevator.Queue[elevator.Floor][elevio.BT_Cab] ||
			!ordersAbove(elevator)
	case config.DOWN:
		return elevator.Queue[elevator.Floor][elevio.BT_HallDown] ||
			elevator.Queue[elevator.Floor][elevio.BT_Cab] ||
			!ordersBelow(elevator)
	case config.STILL:
	}
	return false
}

func motorDirToElevDir(direction elevio.MotorDirection) config.Direction {
	if direction == elevio.MD_Up {
		return config.UP
	} else if direction == elevio.MD_Down {
		return config.DOWN
	} else {
		return config.STILL
	}
}

func printQueue(elevator config.Elev) {
	for button := 0; button < numButtons; button++ {
		for floor := 0; floor < numFloors; floor++ {
			fmt.Println(elevator.Queue[floor][button])
		}
	}
}
