package main

import (
	"encoding/json"
	"fmt"
	"github.com/Millefeuille42/Daemonize"
	"github.com/Millefeuille42/TracimDaemonSDK"
	"net"
)

func sendMessageToSocket(path string, data []byte) error {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(data)
	return err
}

func sendDaemonEvent(path string, event *TracimDaemonSDK.DaemonEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	safeLog(Daemonize.LOG_INFO, fmt.Sprintf("SOCKET: SEND: %s -> %s", event.Type, path))

	return sendMessageToSocket(path, data)
}

func removeConnection(slice []daemonConnection, elementPath string) []daemonConnection {
	for i, c := range slice {
		if c.Path == elementPath {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
