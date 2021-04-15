#!/bin/bash
echo "Velkommen til heis!"
echo "Skriv inn TO heis-simulator port"
read elevPort1
read elevPort2
echo "Skriv inn transmitting port"
read transmitPort
echo "Skriv inn receive port"
read receivePort
cd ..
cd Simulator-v2
gnome-terminal -x ./SimElevatorServer --port=$elevPort1
gnome-terminal -x ./SimElevatorServer --port=$elevPort2
cd ..
cd project-gruppe_19
gnome-terminal -x go run main.go -elev_port=$elevPort2 -transmit_port=$receivePort -receive_port=$transmitPort 
go run main.go -elev_port=$elevPort1 -transmit_port=$transmitPort -receive_port=$receivePort



