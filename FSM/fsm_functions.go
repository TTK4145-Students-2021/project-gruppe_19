package FSM

import (
	"fmt"

	"../config"
	"../driver/elevio"
)

func ordersAbove(elevator config.Elev) bool {
	currentFloor := elevator.Floor
	for i := currentFloor + 1; i < config.NumFloors; i++ {
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
	if elevator.Floor >= 0 {

		for btn := 0; btn < 3; btn++ {
			if elevator.Queue[elevator.Floor][btn] {
				if elevator.Dir == config.UP && (btn == 0) { //makes sure the elevator only stops of the order is in the same direction
					return true
				} else if elevator.Dir == config.DOWN && btn == 1 { //makes sure the elevator only stops of the order is in the same direction
					return true
				} else if btn == 2 { //cab orders will always stop no matter which direction
					return true
				} else if elevator.Floor == 0 || elevator.Floor == (config.NumFloors-1) { //takes care of the edge cases
					return true
				} else if elevator.Dir == config.DOWN && (btn == 0) && !ordersBelow(elevator) {
					return true
				} else if elevator.Dir == config.UP && (btn == 1) && !ordersAbove(elevator) {
					return true
				} else if elevator.Dir == config.STILL {
					return true
				}

			}
		}
		return false
	}
	return false

}

func DeleteOrder(elevator *config.Elev) {
	for i := 0; i < config.NumButtons; i++ {
		elevator.Queue[elevator.Floor][i] = false
	}
}

func DeleteAllOrders(elevator *config.Elev) {
	for btn := 0; btn < config.NumButtons; btn++ {
		for floor := 0; floor < config.NumFloors; floor++ {
			elevator.Queue[floor][btn] = false
			fmt.Println(elevator.Queue[floor][btn])
		}
	}

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
	for button := 0; button < config.NumButtons; button++ {
		for floor := 0; floor < config.NumFloors; floor++ {
			fmt.Println(elevator.Queue[floor][button])
		}
	}
}
