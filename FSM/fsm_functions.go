package FSM

import (
	"fmt"

	"../driver/elevio"
)

type ElevatorState int

const (
	IDLE      = 0
	RUNNING   = 1
	DOOR_OPEN = 2
)

type Direction int

const (
	UP    = 1
	DOWN  = -1
	STILL = 0
)

type Elev struct {
	State ElevatorState
	Dir   Direction
	Floor int
	Queue [numFloors][numButtons]bool
}

type Keypress struct {
	Floor              int
	Btn                elevio.ButtonType
	DesignatedElevator int
	Done               bool
}

func ordersAbove(elevator Elev) bool {
	currentFloor := elevator.Floor
	for i := currentFloor + 1; i < numFloors; i++ {
		if elevator.Queue[i][0] || elevator.Queue[i][1] || elevator.Queue[i][2] {
			return true
		}
	}
	return false
}

func ordersBelow(elevator Elev) bool {
	currentFloor := elevator.Floor
	for i := currentFloor - 1; i > -1; i-- {
		if elevator.Queue[i][0] || elevator.Queue[i][1] || elevator.Queue[i][2] {
			return true
		}
	}
	return false
}

func ordersInFloor(elevator Elev) bool {
	for btn := 0; btn < numButtons; btn++ {
		if elevator.Queue[elevator.Floor][btn] {
			if elevator.Dir == UP && btn == 0 { //makes sure the elevator only stops of the order is in the same direction
				return true
			} else if elevator.Dir == DOWN && btn == 1 { //makes sure the elevator only stops of the order is in the same direction
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

func DeleteOrder(elevator *Elev) {
	deletedFloor := elevator.Floor
	for i := 0; i < numButtons; i++ {
		elevator.Queue[elevator.Floor][i] = false
	}
	fmt.Println("Order deleted at ", deletedFloor)
}

func DeleteAllOrders(elevator *Elev) {
	for btn := 0; btn < numButtons; btn++ {
		for floor := 0; floor < numFloors; floor++ {
			elevator.Queue[floor][btn] = false
			fmt.Println(elevator.Queue[floor][btn])
		}
	}

}

func chooseElevatorDir(elevator Elev) elevio.MotorDirection {
	switch elevator.Dir {
	case STILL:
		if ordersAbove(elevator) {
			return elevio.MD_Up
		} else if ordersBelow(elevator) {
			return elevio.MD_Down
		} else {
			return elevio.MD_Stop
		}
	case UP:
		if ordersAbove(elevator) {
			return elevio.MD_Up
		} else if ordersBelow(elevator) {
			return elevio.MD_Down
		} else {
			return elevio.MD_Stop
		}

	case DOWN:
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

func shouldStop(elevator Elev) bool {
	switch elevator.Dir {
	case UP:
		return elevator.Queue[elevator.Floor][elevio.BT_HallUp] ||
			elevator.Queue[elevator.Floor][elevio.BT_Cab] ||
			!ordersAbove(elevator)
	case DOWN:
		return elevator.Queue[elevator.Floor][elevio.BT_HallDown] ||
			elevator.Queue[elevator.Floor][elevio.BT_Cab] ||
			!ordersBelow(elevator)
	case STILL:
	}
	return false
}

func motorDirToElevDir(direction elevio.MotorDirection) Direction {
	if direction == elevio.MD_Up {
		return UP
	} else if direction == elevio.MD_Down {
		return DOWN
	} else {
		return STILL
	}
}

func printQueue(elevator Elev) {
	for button := 0; button < numButtons; button++ {
		for floor := 0; floor < numFloors; floor++ {
			fmt.Println(elevator.Queue[floor][button])
		}
	}
}
