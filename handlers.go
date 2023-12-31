package main

import (
	"github.com/Millefeuille42/Daemonize"
	"github.com/Millefeuille42/TracimDaemonSDK"
)

func ackHandler(event *TracimDaemonSDK.DaemonEvent) {
	err := sendDaemonEvent(event.Path, &TracimDaemonSDK.DaemonEvent{
		Path: globalConfig.SocketPath,
		Type: TracimDaemonSDK.DaemonAck,
		Data: nil,
	})
	if err != nil {
		safeLog(Daemonize.LOG_ERR, err)
	}
}

func clientAddHandler(event *TracimDaemonSDK.DaemonEvent) {
	err := TracimDaemonSDK.ParseDaemonData(event, &TracimDaemonSDK.DaemonClientData{})
	if err != nil {
		safeLog(Daemonize.LOG_ERR, err)
		return
	}

	connectionsMutex.Lock()
	for i, conn := range connections {
		if conn.Path == event.Data.(*TracimDaemonSDK.DaemonClientData).Path {
			connections[i].isAlive = true
			connections[i].DaemonClientData = *event.Data.(*TracimDaemonSDK.DaemonClientData)
			connectionsMutex.Unlock()
			getAccountInfoHandler(event)
			return
		}
	}
	connections = append(connections, daemonConnection{
		DaemonClientData: *event.Data.(*TracimDaemonSDK.DaemonClientData),
		isAlive:          true,
	})
	connectionsMutex.Unlock()

	getAccountInfoHandler(event)

	broadcastDaemonEvent(&TracimDaemonSDK.DaemonEvent{
		Path: globalConfig.SocketPath,
		Type: TracimDaemonSDK.DaemonClientAdded,
		Data: event.Data,
	})
}

func clientDeleteHandler(event *TracimDaemonSDK.DaemonEvent) {
	connectionsMutex.Lock()
	connections = removeConnection(connections, event.Path)
	connectionsMutex.Unlock()

	broadcastDaemonEvent(&TracimDaemonSDK.DaemonEvent{
		Path: globalConfig.SocketPath,
		Type: TracimDaemonSDK.DaemonClientDeleted,
		Data: event.Data,
	})
}

func pingHandler(event *TracimDaemonSDK.DaemonEvent) {
	err := sendDaemonEvent(event.Path, &TracimDaemonSDK.DaemonEvent{
		Path: globalConfig.SocketPath,
		Type: TracimDaemonSDK.DaemonPong,
		Data: nil,
	})

	if err != nil {
		safeLog(Daemonize.LOG_ERR, err)
	}
}

func pongHandler(event *TracimDaemonSDK.DaemonEvent) {
	connectionsMutex.Lock()
	for i, conn := range connections {
		if conn.Path == event.Path {
			connections[i].isAlive = true
			break
		}
	}
	connectionsMutex.Unlock()
	ackHandler(event)
}

func getClientsHandler(event *TracimDaemonSDK.DaemonEvent) {
	connectionsMutex.Lock()
	err := sendDaemonEvent(event.Path, &TracimDaemonSDK.DaemonEvent{
		Path: globalConfig.SocketPath,
		Type: TracimDaemonSDK.DaemonClients,
		Data: connections,
	})
	connectionsMutex.Unlock()

	if err != nil {
		safeLog(Daemonize.LOG_ERR, err)
	}
}

func doRequestHandler(event *TracimDaemonSDK.DaemonEvent) {
	err := TracimDaemonSDK.ParseDaemonData(event, &TracimDaemonSDK.DaemonDoRequestData{})
	if err != nil {
		safeLog(Daemonize.LOG_ERR, err)
		return
	}

	data := event.Data.(*TracimDaemonSDK.DaemonDoRequestData)
	response, err := s.Request(data.Method, data.Endpoint, data.Body)
	if err != nil {
		safeLog(Daemonize.LOG_ERR, err)
		return
	}

	err = sendDaemonEvent(event.Path, &TracimDaemonSDK.DaemonEvent{
		Path: globalConfig.SocketPath,
		Type: TracimDaemonSDK.DaemonRequestResult,
		Data: TracimDaemonSDK.DaemonRequestResultData{
			Request:    *data,
			StatusCode: response.StatusCode,
			Status:     response.Status,
			Data:       response.DataBytes,
		},
	})

	if err != nil {
		safeLog(Daemonize.LOG_ERR, err)
	}
}

func getAccountInfoHandler(event *TracimDaemonSDK.DaemonEvent) {
	err := sendDaemonEvent(event.Path, &TracimDaemonSDK.DaemonEvent{
		Path: globalConfig.SocketPath,
		Type: TracimDaemonSDK.DaemonAccountInfo,
		Data: TracimDaemonSDK.DaemonAccountInfoData{UserId: s.UserID},
	})

	if err != nil {
		safeLog(Daemonize.LOG_ERR, err)
	}
}
