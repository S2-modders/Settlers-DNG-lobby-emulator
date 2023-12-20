package main

import (
	"net"
	
	"s2dnglobby/config"
	"s2dnglobby/library"
	"s2dnglobby/lobby"
	"s2dnglobby/network"
)

var log = library.GetLogger("Main")
var connChannel = make(chan *net.TCPConn)


func main() {	
	log.Infoln("Starting S2 DNG Lobby Server")

	lobby.InitLobby()

	var addr = net.TCPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: config.SERVER_PORT,
	}

	log.Infoln("Listening on", addr.String())

	listener, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		log.Panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Errorln("Failed to accept TCP connection")
			continue
		}
		conn.SetNoDelay(true)
		conn.SetReadBuffer(1024)

		go network.HandleConnection(conn)
	}

	//var data = make([]byte, 1024)
	/*
	for {
		var header = new(packages.Header)

		if err := binary.Read(conn, binary.LittleEndian, header); err != nil {
			//panic(err)
			fmt.Println(err)
		}

		fmt.Println("got header:", header)
		fmt.Println(
			hex.EncodeToString(header.Magic[:]), 
			header.HeaderType, 
			header.PayloadSize,
			hex.EncodeToString(header.PayloadChecksum[:]),
		)

		break
	}
	*/

	/*
	for {
		n, err := conn.Read(data)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			} else {
				panic(err)
			}
		}

		fmt.Printf("got %d bytes: ", n)

		var payload = data[:n]
		fmt.Println(hex.EncodeToString(payload))
	}
	*/
}
