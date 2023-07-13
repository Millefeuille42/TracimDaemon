package main

import (
	"encoding/json"
	"github.com/Millefeuille42/TracimAPI/session"
	"github.com/Millefeuille42/TracimDaemonSDK"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var socketPath string
var socket net.Listener
var connections = make([]string, 0)

func removeConnection(slice []string, element string) []string {
	for i, c := range connections {
		if c == element {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func connectedHandler(s *session.Session, TLM *session.TracimLiveMessage) {
	log.Print("Connected")
}

func messageHandler(s *session.Session, TLM *session.TracimLiveMessage) {
	log.Printf("Received event: %s\n", TLM.DataParsed.EventType)
	for _, connPath := range connections {
		conn, err := net.Dial("unix", connPath)
		if err != nil {
			log.Print(err)
			continue
		}
		_, err = conn.Write([]byte(TLM.Data))
		if err != nil {
			log.Print(err)
		}
		conn.Close()
	}
}

func errorHandler(s *session.Session, TLM *session.TracimLiveMessage) {
	log.Print("Error: " + TLM.Data)
}

func handleSigTerm() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		_ = os.Remove(socketPath)
		os.Exit(1)
	}()
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
			log.Printf("SOCKET: %s\n", buf[:n])
			message := TracimDaemonSDK.DaemonSubscriptionEvent{}
			err = json.Unmarshal(buf[:n], &message)
			if err != nil {
				log.Print(err)
				return
			}

			switch message.Action {
			case TracimDaemonSDK.DaemonSubscriptionActionAdd:
				connections = append(connections, message.Path)
			case TracimDaemonSDK.DaemonSubscriptionActionDelete:
				connections = removeConnection(connections, message.Path)
			}
		}(conn)
	}
}

func prepareTracimClient() *session.Session {
	s := session.New(os.Getenv("TRACIM_URL"))
	s.SetCredentials(session.Credentials{
		Username: os.Getenv("TRACIM_USERNAME"),
		Mail:     os.Getenv("TRACIM_MAIL"),
		Password: os.Getenv("TRACIM_PASSWORD"),
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
	socketPath = os.Getenv("TRACIM_SOCKET_PATH")
	handleSigTerm()

	var err error
	socket, err = net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer socket.Close()

	go listenConnections()

	s := prepareTracimClient()
	s.ListenEvents(os.Getenv("TRACIM_USER_ID"))
}
