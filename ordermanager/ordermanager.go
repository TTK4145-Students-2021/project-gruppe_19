package ordermanager

import (
	"fmt"
	"math"

	"../config"
	"../driver/elevio"
)

const numElev = 3

func costFunc(id string, orderMap map[string]config.Elev, orderFloor int) string { //TODO: some less basic cost function maybe?, works OK though.
	closestDist := 1000.0 //just something large
	bestElevID := id
	for id, elev := range orderMap {
		if math.Abs(float64(elev.Floor-orderFloor)) < float64(closestDist) {
			closestDist = math.Abs(float64(elev.Floor - orderFloor))
			bestElevID = id
		}
	}
	println("closestDist: ", closestDist)
	return bestElevID

}

func printMap(orderMap map[string]config.Elev) {
	for id, elev := range orderMap {
		println("ElevatorID: ", id)
		println("Elevator Floor: ", elev.Floor)
	}
}

func OrderMan(orderChan config.OrderChannels, elevChan config.ElevChannels, id string, elev *config.Elev) {

	orderMap := make(map[string]config.Elev)
	orderMap[id] = *elev //insert this elevator into map with corresponding ID
	for {
		select {
		case incomingOrder := <-orderChan.DelegateOrder:

			if incomingOrder.Button == elevio.BT_Cab { //cab orders are handled by the ordered elevator, always
				orderChan.ExtOrder <- incomingOrder
			} else {
				printMap(orderMap)
				orderFloor := incomingOrder.Floor
				bestElevID := costFunc(id, orderMap, orderFloor)

				if bestElevID == id { //if the chosen best elevator is this one, just send it to FSM
					orderChan.ExtOrder <- incomingOrder
				} else { //if its one of the others, send it over the net
					orderChan.SendOrder <- incomingOrder
					orderChan.ExternalID <- bestElevID
				}

				fmt.Println("selected elev: ", bestElevID)
			}

		case incMap := <-elevChan.MapChan: //update map
			for incId, incElev := range incMap { //TODO: maps give have floor == 0 when elevator starts between floors, needs fixing
				orderMap[incId] = incElev
			}
		}
	}

}
