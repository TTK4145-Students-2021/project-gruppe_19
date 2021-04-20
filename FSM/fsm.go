package FSM

import (
	"time"

	"../config"
	"../driver/elevio"
)

const elevSendInterval = 100 * time.Millisecond //interval for how often we send the current elevator over elevatorchannel

const timerTime = 4 //how many seconds is needed to make the engine go into error
var engineErrorTimer = time.NewTimer(timerTime * time.Second)

var dir elevio.MotorDirection //the direction of the elevator

//If the elevator starts between floor, this function takes them to a floor. Only called once.
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

//If the motor has stopped, this function will be called to try and restart it
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

//Removes all lampts at current floor
func removeButtonLamps(elevator config.Elev) {
	elevio.SetButtonLamp(elevio.BT_Cab, elevator.Floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, elevator.Floor, false)
	elevio.SetButtonLamp(elevio.BT_HallUp, elevator.Floor, false)
}

//Runs the finite state machine of a single elevator
func Fsm(elevChan config.ElevChannels, elevator *config.Elev, drvChan config.DriverChannels) {
	engineErrorTimer.Stop()
	for {
		switch elevator.State {
		case config.IDLE:
			if ordersAbove(*elevator) {
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
			doorTimer := time.NewTimer(2 * time.Second) //door is open i 2 seconds
			<-doorTimer.C
			elevio.SetDoorOpenLamp(false)
			engineErrorTimer.Reset(timerTime * time.Second)

		case config.ERROR: //motor is out. Try to restart.
			println("In ERROR state. Trying to restart...")
			tryRestartMotor(elevator, drvChan)

		case config.OBSTRUCTED: //waits until obstruction is pressed again to resume
			engineErrorTimer.Stop()
			elevio.SetDoorOpenLamp(true)
			dir = elevio.MD_Stop
			elevio.SetMotorDirection(dir)
			elevio.SetFloorIndicator(elevator.Floor)
			<-drvChan.DrvObstr
			elevio.SetDoorOpenLamp(false)
			elevator.State = config.IDLE
		}
	}

}

//Function which receives input from elevator panel and handles the orders and events accordingly. Main loop of program.
func InternalControl(drvChan config.DriverChannels, orderChan config.OrderChannels, elevChan config.ElevChannels, elevator *config.Elev) {
	FsmInit(elevator, drvChan)

	for {
		select {
		case floor := <-drvChan.DrvFloors: //Sensor senses a new floor
			elevator.Floor = floor
			engineErrorTimer.Reset(timerTime * time.Second)

		case drvOrder := <-drvChan.DrvButtons: // a new button is pressed on this elevator
			orderChan.DelegateOrder <- drvOrder //order is sent to ordermanager
			elevio.SetButtonLamp(drvOrder.Button, drvOrder.Floor, true)

		case ExtOrder := <-orderChan.ExtOrder:
			elevator.Queue[ExtOrder.Floor][int(ExtOrder.Button)] = true //this elevator takes the order
			elevio.SetButtonLamp(ExtOrder.Button, ExtOrder.Floor, true)

		case <-drvChan.DoorsOpen: //Doors open. Order is completed and sent to the other elevators to turn off order light
			elevChan.Elevator <- *elevator
			order1, order2 := getOrder(elevator)
			deleteOrder(elevator)
			if order1.Floor != -1 { //only valid orders get sent
				orderChan.CompletedOrder <- order1
			}
			if order2.Floor != -1 {
				orderChan.CompletedOrder <- order2
			}

		case <-drvChan.DrvStop: // Stop button is pressed. Will start again after 3 seconds or go into error state
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

		case <-drvChan.DrvObstr: //Send elevator to OBSTRUCTED state. Opens doors until obstruction is pressed again.
			elevator.State = config.OBSTRUCTED

		case <-time.After(elevSendInterval): //Updates elevator channel with current elevator every *elevSendInterval*
			elevChan.Elevator <- *elevator

		case <-engineErrorTimer.C: //If timer runs out the motor has stopped. Try to restart
			if !ordersAbove(*elevator) || !ordersBelow(*elevator) || !ordersInFloor(*elevator) { //no orders are left
				println("motor stopped")
				elevator.State = config.ERROR

			} else {
				engineErrorTimer.Stop()
			}
		}

	}
}
