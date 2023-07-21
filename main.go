package main

import (
	"encoding/json"
	"fmt"
	"github.com/Millefeuille42/Daemonize"
	"github.com/Millefeuille42/TracimAPI/session"
	"github.com/Millefeuille42/TracimDaemonSDK"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type daemonConnection struct {
	TracimDaemonSDK.DaemonClientData
	isAlive bool
}

var socket net.Listener
var connections = make([]daemonConnection, 0)
var s *session.Session
var daemonizer *Daemonize.Daemonizer = nil

var connectionsMutex = sync.Mutex{}
var logMutex = sync.Mutex{}

func safeLog(severity Daemonize.Severity, v ...any) {
	logMutex.Lock()
	defer logMutex.Unlock()
	if daemonizer == nil {
		log.Print(v...)
		return
	}
	daemonizer.Log(severity, v...)
}

func broadcastDaemonEvent(e *TracimDaemonSDK.DaemonEvent) {
	connectionsMutex.Lock()
	for _, connData := range connections {
		err := sendDaemonEvent(connData.Path, e)
		if err != nil {
			safeLog(Daemonize.LOG_ERR, err)
		}
	}
	connectionsMutex.Unlock()
}

func sendPings() {
	connectionsMutex.Lock()
	oldConnections := connections
	for _, conn := range oldConnections {
		if !conn.isAlive {
			removeConnection(connections, conn.Path)
			safeLog(Daemonize.LOG_INFO, fmt.Sprintf("PING: Removed %s due to inactivty", conn.Path))
		}
	}
	for i := range connections {
		connections[i].isAlive = false
	}
	connectionsMutex.Unlock()

	broadcastDaemonEvent(&TracimDaemonSDK.DaemonEvent{
		Path: globalConfig.SocketPath,
		Type: TracimDaemonSDK.DaemonPing,
		Data: nil,
	})
}

func connectedHandler(s *session.Session, TLM *session.TracimLiveMessage) {
	safeLog(Daemonize.LOG_INFO, "TRACIM: Connected")
}

func messageHandler(s *session.Session, TLM *session.TracimLiveMessage) {
	safeLog(Daemonize.LOG_INFO, fmt.Sprintf("TRACIM: RECV: %s\n", TLM.DataParsed.EventType))
	broadcastDaemonEvent(&TracimDaemonSDK.DaemonEvent{
		Path: globalConfig.SocketPath,
		Type: TracimDaemonSDK.DaemonTracimEvent,
		Data: TLM.Data,
	})
}

func errorHandler(s *session.Session, TLM *session.TracimLiveMessage) {
	safeLog(Daemonize.LOG_ERR, "TRACIM: ERROR: "+TLM.Data)
}

func startPingRoutine() {
	pingTicker := time.NewTicker(time.Second * 30)
	defer pingTicker.Stop()
	for {
		select {
		case <-pingTicker.C:
			sendPings()
		}
	}
}

func listenConnections() {
	for {
		conn, err := socket.Accept()
		if err != nil {
			safeLog(Daemonize.LOG_ERR, err)
			continue
		}
		go func(conn net.Conn) {
			defer conn.Close()
			buf := make([]byte, 4096)

			n, err := conn.Read(buf)
			if err != nil {
				safeLog(Daemonize.LOG_ERR, err)
				return
			}
			message := TracimDaemonSDK.DaemonEvent{}
			err = json.Unmarshal(buf[:n], &message)
			if err != nil {
				safeLog(Daemonize.LOG_ERR, err)
				return
			}

			safeLog(Daemonize.LOG_INFO, fmt.Sprintf("SOCKET: RECV: %s -> %s\n", message.Type, message.Path))

			switch message.Type {
			case TracimDaemonSDK.DaemonClientAdd:
				clientAddHandler(&message)
			case TracimDaemonSDK.DaemonClientDelete:
				clientDeleteHandler(&message)
			case TracimDaemonSDK.DaemonGetClients:
				getClientsHandler(&message)
			case TracimDaemonSDK.DaemonDoRequest:
				doRequestHandler(&message)
			case TracimDaemonSDK.DaemonGetAccountInfo:
				getAccountInfoHandler(&message)
			case TracimDaemonSDK.DaemonPing:
				pingHandler(&message)
			case TracimDaemonSDK.DaemonPong:
				pongHandler(&message)
			case TracimDaemonSDK.DaemonAck:
			default:
				ackHandler(&message)
			}
		}(conn)
	}
}

func prepareTracimClient() *session.Session {
	s = session.New(globalConfig.Tracim.Url)
	s.SetCredentials(session.Credentials{
		Username: globalConfig.Tracim.Username,
		Mail:     globalConfig.Tracim.Mail,
		Password: globalConfig.Tracim.Password,
	})

	err := s.Auth()
	if err != nil {
		safeLog(Daemonize.LOG_EMERG, err)
		os.Exit(1)
	}

	s.TLMSubscribe(session.TLMError, errorHandler)
	s.TLMSubscribe(session.TLMConnected, connectedHandler)
	s.TLMSubscribe(session.TLMMessage, messageHandler)

	return s
}

func startProcess() {
	safeLog(Daemonize.LOG_INFO, "Started")
	defer os.Remove(globalConfig.SocketPath)
	_ = os.Remove(globalConfig.SocketPath)

	var err error
	socket, err = net.Listen("unix", globalConfig.SocketPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer socket.Close()

	go listenConnections()

	prepareTracimClient()

	go startPingRoutine()
	s.ListenEvents()
}

func main() {
	setGlobalConfig()

	if len(os.Args) > 1 && os.Args[1] == "-p" {
		startProcess()
		os.Exit(0)
	}

	var err error = nil
	daemonizer, err = Daemonize.NewDaemonizer()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer daemonizer.Close()

	pid, err := daemonizer.Daemonize(nil)
	if err != nil {
		log.Fatal(err)
	}
	if pid != 0 {
		log.Print(pid)
		os.Exit(0)
	}

	pattern := fmt.Sprintf("master_%s_*.log", time.Now().Format(time.RFC3339))
	err = daemonizer.AddTempFileLogger(configDir+"log", pattern, os.Args[0], log.LstdFlags)
	if err != nil {
		log.Fatal(err)
		return
	}

	startProcess()
}
