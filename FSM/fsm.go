package FSM

import (
	"time"

	"../config"
	"../driver/elevio"
)

const elevSendInterval = 100 * time.Millisecond //timer for how often we send the current elevator over elevatorchannel
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
			for i := 0; i < config.NumFloors; i++ {
				for j := elevio.BT_HallUp; j < config.NumButtons; j++ {
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

	for !success {
		if elevator.Floor < (config.NumFloors - 2) { //last sensed floor
			elevio.SetMotorDirection(elevio.MD_Up) //just sends it some safe direction. Not trying for optimality
		} else {
			elevio.SetMotorDirection(elevio.MD_Down)
		}
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

func Fsm(elevChan config.ElevChannels, elevator *config.Elev, drvChan config.DriverChannels) {
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
			elevator.State = config.IDLE
			drvChan.DoorsOpen <- elevator.Floor
			removeButtonLamps(*elevator)
			doorTimer := time.NewTimer(2 * time.Second)
			<-doorTimer.C
			elevio.SetDoorOpenLamp(false)
			engineErrorTimer.Reset(timerTime * time.Second)

		case config.ERROR:
			println("In ERROR state. Trying to restart...")
			tryRestartMotor(elevator, drvChan)

		}
	}

}

func InternalControl(drvChan config.DriverChannels, orderChan config.OrderChannels, elevChan config.ElevChannels, elevator *config.Elev) {
	FsmInit(elevator, drvChan)

	for {
		select {
		case floor := <-drvChan.DrvFloors: //Sensor senses a new floor
			//FsmUpdateFloor(floor, elevator)
			elevator.Floor = floor
			engineErrorTimer.Reset(timerTime * time.Second)

		case drvOrder := <-drvChan.DrvButtons: // a new button is pressed on this elevator
			orderChan.DelegateOrder <- drvOrder
			elevio.SetButtonLamp(drvOrder.Button, drvOrder.Floor, true)

		case ExtOrder := <-orderChan.ExtOrder:
			elevator.Queue[ExtOrder.Floor][int(ExtOrder.Button)] = true
			elevio.SetButtonLamp(ExtOrder.Button, ExtOrder.Floor, true)

		case <-drvChan.DoorsOpen:
			elevChan.Elevator <- *elevator
			order1, order2 := getOrder(elevator)
			deleteOrder(elevator)
			if order1.Floor != -1 {
				orderChan.CompletedOrder <- order1
			} else if order2.Floor != -1 {
				orderChan.CompletedOrder <- order2
			} else {
				//do nothing
			}

		case <-drvChan.DrvStop: //TODO: check if this is the wanted functionality
			if elevator.State == config.IDLE {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetStopLamp(true)
				println("Starting again in 3 seconds")
				waited := false
				stopTimer := time.NewTimer(3 * time.Second)
				for !waited {
					select {
					case <-stopTimer.C:
						waited = true
						elevio.SetStopLamp(false)
					}
				}

			} else {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetStopLamp(true)
				time.Sleep(3 * time.Second) //will go to error state
			}

		case <-drvChan.DrvObstr: //TODO: add some functionality here?
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
