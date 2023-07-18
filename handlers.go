package main

import (
	"github.com/Millefeuille42/TracimDaemonSDK"
	"log"
)

func ackHandler(message *TracimDaemonSDK.DaemonEvent) {
	err := sendDaemonEvent(message.Path, &TracimDaemonSDK.DaemonEvent{
		Path:   socketPath,
		Action: TracimDaemonSDK.DaemonAck,
		Data:   nil,
	})
	if err != nil {
		log.Print(err)
	}
}

func subscriptionActionAddHandler(message *TracimDaemonSDK.DaemonEvent) {
	connectionsMutex.Lock()
	connections = append(connections, daemonConnection{
		path:    message.Path,
		isAlive: true,
	})
	connectionsMutex.Unlock()

	err := sendDaemonEvent(message.Path, &TracimDaemonSDK.DaemonEvent{
		Path:   socketPath,
		Action: TracimDaemonSDK.DaemonAccountInfo,
		Data:   userId,
	})

	if err != nil {
		log.Print(err)
	}
}

func subscriptionActionDeleteHandler(message *TracimDaemonSDK.DaemonEvent) {
	connectionsMutex.Lock()
	connections = removeConnection(connections, message.Path)
	connectionsMutex.Unlock()
	ackHandler(message)
}

func pingHandler(message *TracimDaemonSDK.DaemonEvent) {
	err := sendDaemonEvent(message.Path, &TracimDaemonSDK.DaemonEvent{
		Path:   socketPath,
		Action: TracimDaemonSDK.DaemonPong,
		Data:   nil,
	})

	if err != nil {
		log.Print(err)
	}
}

func pongHandler(message *TracimDaemonSDK.DaemonEvent) {
	connectionsMutex.Lock()
	for i, conn := range connections {
		if conn.path == message.Path {
			connections[i].isAlive = true
			break
		}
	}
	connectionsMutex.Unlock()
	ackHandler(message)
}
