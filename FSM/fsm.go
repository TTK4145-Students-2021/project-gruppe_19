package FSM

import (
	"time"

	"../config"
	"../driver/elevio"
)

const numFloors = 4
const numButtons = 3
const elevSendInterval = 1000 * time.Millisecond //timer for how often we send the current elevator over elevatorchannel
const timerTime = 4

var engineErrorTimer = time.NewTimer(timerTime * time.Second)

var dir elevio.MotorDirection

func FsmInit(elevator *config.Elev, drvChan config.DriverChannels) {
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(elevio.MD_Down)
	for {
		select {
		case floor := <-drvChan.DrvFloors:
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetFloorIndicator(floor)
			elevator.Floor = floor
			for i := 0; i < numFloors; i++ {
				for j := elevio.BT_HallUp; j < numButtons; j++ {
					elevio.SetButtonLamp(j, i, false)
				}
			}
			return
		}
	}

}

func FsmUpdateFloor(newFloor int, elevator *config.Elev) { //hvordan dette skal gjøres igjen
	elevator.Floor = newFloor
	engineErrorTimer.Reset(timerTime * time.Second)

}

func tryRestartMotor(elevator *config.Elev, drvChan config.DriverChannels) {
	success := false

	if elevator.Floor < (numFloors - 2) { //last sensed floor
		elevio.SetMotorDirection(elevio.MD_Up)
	} else {
		elevio.SetMotorDirection(elevio.MD_Down)
	}

	for !success {
		select {
		case sensedNewFloor := <-drvChan.DrvFloors:
			elevator.Floor = sensedNewFloor
			elevator.State = config.IDLE
			success = true
			engineErrorTimer.Reset(timerTime * time.Second)
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetStopLamp(false)
			println("Restart success!")
		}

	}
	return

}

func removeButtonLamps(elevator config.Elev) {
	elevio.SetButtonLamp(elevio.BT_Cab, elevator.Floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, elevator.Floor, false)
	elevio.SetButtonLamp(elevio.BT_HallUp, elevator.Floor, false)
}

func Fsm(doorsOpen chan<- int, elevChan config.ElevChannels, elevator *config.Elev, drvChan config.DriverChannels) {
	engineErrorTimer.Stop()
	for {
		switch elevator.State {
		case config.IDLE:
			if ordersAbove(*elevator) {
				//println("order above,going up, current Floor: ", Floor)
				dir = elevio.MD_Up
				elevator.Dir = motorDirToElevDir(dir)
				elevio.SetMotorDirection(dir)
				elevator.State = config.RUNNING
				engineErrorTimer.Reset(timerTime * time.Second)

			}
			if ordersBelow(*elevator) {
				dir = elevio.MD_Down
				elevator.Dir = motorDirToElevDir(dir)
				elevio.SetMotorDirection(dir)
				elevator.State = config.RUNNING
				engineErrorTimer.Reset(timerTime * time.Second)
			}
			elevator.Floor = elevio.GetFloor()
			if ordersInFloor(*elevator) {
				elevator.State = config.DOOR_OPEN
			}
			engineErrorTimer.Reset(timerTime * time.Second)

		case config.RUNNING:
			elevator.Floor = elevio.GetFloor()
			if ordersInFloor(*elevator) {
				dir = elevio.MD_Stop
				elevator.Dir = motorDirToElevDir(dir)
				elevio.SetMotorDirection(dir)
				elevator.State = config.DOOR_OPEN

			}
		case config.DOOR_OPEN:
			engineErrorTimer.Stop()
			elevio.SetDoorOpenLamp(true)
			dir = elevio.MD_Stop
			elevio.SetMotorDirection(dir)
			elevio.SetFloorIndicator(elevator.Floor)
			DeleteOrder(elevator)
			elevator.State = config.IDLE
			doorsOpen <- elevator.Floor
			timer1 := time.NewTimer(2 * time.Second)
			<-timer1.C
			elevio.SetDoorOpenLamp(false)
			removeButtonLamps(*elevator)
			engineErrorTimer.Reset(timerTime * time.Second)

		case config.ERROR:
			println("In ERROR state. Trying to restart...")
			tryRestartMotor(elevator, drvChan)

		}
	}

}

// InternalControl .. Responsible for internal control of a single elevator
func InternalControl(drvChan config.DriverChannels, orderChan config.OrderChannels, elevChan config.ElevChannels, elevator *config.Elev) {
	FsmInit(elevator, drvChan)

	for {
		select {
		case floor := <-drvChan.DrvFloors: //Sensor senses a new floor
			FsmUpdateFloor(floor, elevator)

		case drvOrder := <-drvChan.DrvButtons: // a new button is pressed on this elevator
			orderChan.DelegateOrder <- drvOrder

			/*
				elevator.Queue[drvOrder.Floor][int(drvOrder.Button)] = true
				elevio.SetButtonLamp(drvOrder.Button, drvOrder.Floor, true)*/
		case ExtOrder := <-orderChan.ExtOrder:
			elevator.Queue[ExtOrder.Floor][int(ExtOrder.Button)] = true
			elevio.SetButtonLamp(ExtOrder.Button, ExtOrder.Floor, true)

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
			elevChan.Elevator <- *elevator

		case <-drvChan.DrvStop: //TODO: add some functionality here
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetStopLamp(true)

		case <-drvChan.DrvObstr: //TODO: add some functionality here
			elevator.State = config.DOOR_OPEN

		case <-time.After(elevSendInterval): //Updates elevator channel with current elevator every *elevSendInterval*
			elevChan.Elevator <- *elevator

		case <-engineErrorTimer.C:
			if !ordersAbove(*elevator) || !ordersBelow(*elevator) || !ordersInFloor(*elevator) { //no orders are left
				println("motor stopped")
				elevator.State = config.ERROR

			} else {
				engineErrorTimer.Stop()
			}
		}

	}
}
