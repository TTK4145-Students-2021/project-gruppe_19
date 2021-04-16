package ordermanager

import (
	"fmt"
	"math"

	"../config"
)

const numElev = 3

func costFunc(id string, orderMap map[string]config.Elev, orderFloor int) string {
	closestDist := 1000.0
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

func OrderMan(orderChan config.OrderChannels, elevChan config.ElevChannels, mapChan chan map[string]config.Elev, id string, elev *config.Elev) {

	orderMap := make(map[string]config.Elev)
	orderMap[id] = *elev //insert this elevator into map with corresponding ID
	for {
		select {
		case incomingOrder := <-orderChan.DelegateOrder:

			//TODO:få også ID-en og mapsa modulært
			orderFloor := incomingOrder.Floor

			bestElevID := costFunc(id, orderMap, orderFloor)

			for id, _ := range orderMap {
				println("id in ordermap: ", id)
			}

			if bestElevID == id {
				orderChan.ExtOrder <- incomingOrder
			} else {
				orderChan.SendOrder <- incomingOrder
				orderChan.ExternalID <- bestElevID
			}

			fmt.Println("selected elev: ", bestElevID)

		case incMap := <-mapChan:
			//TODO:må finne ut hvorfor mapsa blir forskjellig for hver heis
			for incId, incElev := range incMap {
				orderMap[incId] = incElev
			}
		}
	}

}

//Start groutine to handle order

//Do cost calculation (send to cost function)

//If itself: send to fsm and broadcast
//Else: send to correct elevator, and monitor.

//Wait for order broadcast

//Start goroutine to deal with broadcast

//If taking order: send to fsm (wait 1 sec) and broadcast order accept
//Else: Wait for order accept, then wait for order finish
