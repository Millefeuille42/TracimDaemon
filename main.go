package main

import (
	"encoding/json"
	"github.com/Millefeuille42/TracimAPI/session"
	"github.com/Millefeuille42/TracimDaemonSDK"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type daemonConnection struct {
	TracimDaemonSDK.DaemonClientData
	isAlive bool
}

var socket net.Listener
var connections = make([]daemonConnection, 0)
var connectionsMutex = sync.Mutex{}
var s *session.Session

func broadcastDaemonEvent(e *TracimDaemonSDK.DaemonEvent) {
	connectionsMutex.Lock()
	for _, connData := range connections {
		err := sendDaemonEvent(connData.Path, e)
		if err != nil {
			log.Print(err)
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
			log.Printf("PING: Removed %s due to inactivty", conn.Path)
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
	log.Print("TRACIM: Connected")
}

func messageHandler(s *session.Session, TLM *session.TracimLiveMessage) {
	log.Printf("TRACIM: RECV: %s\n", TLM.DataParsed.EventType)
	broadcastDaemonEvent(&TracimDaemonSDK.DaemonEvent{
		Path: globalConfig.SocketPath,
		Type: TracimDaemonSDK.DaemonTracimEvent,
		Data: TLM.Data,
	})
}

func errorHandler(s *session.Session, TLM *session.TracimLiveMessage) {
	log.Print("TRACIM: ERROR: " + TLM.Data)
}

func handleSigTerm() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		_ = os.Remove(globalConfig.SocketPath)
		os.Exit(1)
	}()
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
			log.Print(err)
			continue
		}
		go func(conn net.Conn) {
			defer conn.Close()
			buf := make([]byte, 4096)

			n, err := conn.Read(buf)
			if err != nil {
				log.Print(err)
				return
			}
			message := TracimDaemonSDK.DaemonEvent{}
			err = json.Unmarshal(buf[:n], &message)
			if err != nil {
				log.Print(err)
				return
			}

			log.Printf("SOCKET: RECV: %s -> %s\n", message.Type, message.Path)

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
		log.Fatal(err)
		return nil
	}

	s.TLMSubscribe(session.TLMError, errorHandler)
	s.TLMSubscribe(session.TLMConnected, connectedHandler)
	s.TLMSubscribe(session.TLMMessage, messageHandler)

	return s
}

func main() {
	setGlobalConfig()
	handleSigTerm()

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
