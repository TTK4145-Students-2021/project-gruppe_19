package FSM

import (
	"fmt"
	"time"

	"../config"
	"../driver/elevio"
)

/*
func goToFloor(floorRequest int, elevatorState <-chan int) {
    currentFloor := <- elevatorState

    fmt.Println(currentFloor)

}
*/

const numFloors = 4
const numButtons = 3

var dir elevio.MotorDirection

func FsmInit(elevator *config.Elev) {

	// Needs to start in a well-defined state
	for elevator.Floor = elevio.GetFloor(); elevator.Floor < 0; elevator.Floor = elevio.GetFloor() {
		elevio.SetMotorDirection(elevio.MD_Up)
		time.Sleep(1 * time.Second)
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
	fmt.Println("FSM initialized!")
}

func FsmUpdateFloor(newFloor int, elevator *config.Elev) { //hvordan dette skal gjøres igjen
	elevator.Floor = newFloor
}

func removeButtonLamps(elevator config.Elev) {
	elevio.SetButtonLamp(elevio.BT_Cab, elevator.Floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, elevator.Floor, false)
	elevio.SetButtonLamp(elevio.BT_HallUp, elevator.Floor, false)
}

func Fsm(doorsOpen chan<- int, elevChan config.ElevChannels, elevator *config.Elev) {

	for {
		switch elevator.State {
		case config.IDLE:
			if ordersAbove(*elevator) {
				//println("order above,going up, current Floor: ", Floor)
				dir = elevio.MD_Up
				elevator.Dir = motorDirToElevDir(dir)
				elevio.SetMotorDirection(dir)
				elevator.State = config.RUNNING

			}
			if ordersBelow(*elevator) {
				//println("order below, going down, current Floor: ", Floor)
				dir = elevio.MD_Down
				elevator.Dir = motorDirToElevDir(dir)
				elevio.SetMotorDirection(dir)
				elevator.State = config.RUNNING
			}
			if ordersInFloor(*elevator) {
				//println("order below, going down, current Floor: ", Floor)
				elevator.State = config.DOOR_OPEN
			}
			elevChan.Elevator <- *elevator
		case config.RUNNING:
			if ordersInFloor(*elevator) { // this is the problem : the floor is being kept constant at e.g. 2 while its moving
				dir = elevio.MD_Stop
				elevator.Dir = motorDirToElevDir(dir)
				elevio.SetMotorDirection(dir)
				elevator.State = config.DOOR_OPEN
			}
			elevChan.Elevator <- *elevator

		case config.DOOR_OPEN:
			printQueue(*elevator)
			elevio.SetDoorOpenLamp(true)
			dir = elevio.MD_Stop
			elevio.SetMotorDirection(dir)
			DeleteOrder(elevator)
			elevator.State = config.IDLE
			doorsOpen <- elevator.Floor
			timer1 := time.NewTimer(2 * time.Second)
			<-timer1.C
			elevio.SetDoorOpenLamp(false)
			removeButtonLamps(*elevator)
			println("DOOR CLOSE")
			elevChan.Elevator <- *elevator

		}
	}

}

// InternalControl .. Responsible for internal control of a single elevator
func InternalControl(drvChan config.DriverChannels, orderChan config.OrderChannels, elevChan config.ElevChannels, elevator *config.Elev) {
	FsmInit(elevator)
	for {
		select {
		case floor := <-drvChan.DrvFloors: //Sensor senses a new floor
			//println("updating floor:", floor)
			FsmUpdateFloor(floor, elevator)

		case drvOrder := <-drvChan.DrvButtons: // a new button is pressed on this elevator
			orderChan.DelegateOrder <- drvOrder //Delegate this order
			fmt.Println("New order delegated")
			fmt.Println(drvOrder)
			/*
				elevator.Queue[drvOrder.Floor][int(drvOrder.Button)] = true
				elevio.SetButtonLamp(drvOrder.Button, drvOrder.Floor, true)*/
		case ExtOrder := <-orderChan.ExtOrder:
			//AddOrder(ExtOrder)
			fmt.Println("New order externally")
			elevator.Queue[ExtOrder.Floor][int(ExtOrder.Button)] = true
			elevio.SetButtonLamp(ExtOrder.Button, ExtOrder.Floor, true)
			elevChan.Elevator <- *elevator

		case floor := <-drvChan.DoorsOpen:

			order_OutsideUp_Completed := elevio.ButtonEvent{
				Floor:  floor,
				Button: elevio.BT_HallUp,
			}
			order_OutsideDown_Completed := elevio.ButtonEvent{
				Floor:  floor,
				Button: elevio.BT_HallDown,
			}
			order_Inside_Completed := elevio.ButtonEvent{
				Floor:  floor,
				Button: elevio.BT_Cab,
			}
			drvChan.CompletedOrder <- order_OutsideUp_Completed
			drvChan.CompletedOrder <- order_OutsideDown_Completed
			drvChan.CompletedOrder <- order_Inside_Completed

		case <-drvChan.DrvStop:
			elevio.SetMotorDirection(elevio.MD_Stop)
			time.Sleep(3 * time.Second)

		case <-drvChan.DrvObstr:
			elevator.State = config.DOOR_OPEN

		}

	}
}
