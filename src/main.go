package main

import (
	"net"
	"os"
	"slices"

	"s2dnglobby/config"
	"s2dnglobby/library"
	"s2dnglobby/lobby"
	"s2dnglobby/netbridge"
	"s2dnglobby/network"
)

var log = library.GetLogger("Main")

func main() {
	log.Infoln("Starting S2 DNG Lobby Server")

	runLocal := slices.Contains(os.Args, "--local")

	if err := library.DepsCheck(runLocal); err != nil {
		log.Fatalln(err)
		return
	}

	netbridge.InitBridgeController()
	lobby.InitLobby()

	addr := net.TCPAddr{
		IP: net.IPv4zero,
		Port: config.SERVER_PORT,
	}

	log.Infoln("Listening on", addr.String())

	listener, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Errorln("Failed to accept TCP connection")
			continue
		}
		conn.SetNoDelay(true)
		conn.SetReadBuffer(4096)

		go network.HandleConnection(conn)
	}
}
