package network

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"s2dnglobby/config"
	"s2dnglobby/library"
	"s2dnglobby/lobby"
	"s2dnglobby/packages"
)

var log = library.GetLogger("ConnHandler")


func HandleConnection(conn *net.TCPConn) {
	defer conn.Close()
	log.Debugln("Got new connection from", conn.RemoteAddr())

	if err := handleHandshake(conn); err != nil {
		log.Errorln("Handshake failed:", err)
		return
	}

	for {
		header, err := getHeader(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				notifyUserLoggedOut(conn)
			} else {
				log.Errorln(err)
			}
			break
		}

		if header.HeaderType == packages.Ping {
			tbuf := make([]byte, header.PayloadSize)
			conn.Read(tbuf)

			log.Debugln(" <-- Ping:", hex.EncodeToString(tbuf))
			continue
		}
		if header.HeaderType != packages.ApplicationMessage {
			log.Errorln("Got unexpected HeaderType:", header.HeaderType)
			continue
		}

		payload := make([]byte, header.PayloadSize)

		if _, err := conn.Read(payload); err != nil {
			log.Errorln("Failed to recv package payload")
			continue
		}

		//log.Debugln(hex.EncodeToString(payload))

		payloadBuf := bytes.NewBuffer(payload)
		msgHeader := new(packages.MsgHeader)

		if err := binary.Read(payloadBuf, binary.LittleEndian, msgHeader); err != nil {
			log.Errorln("Failed to parse MsgHeader", err)
			continue
		}
		if err := msgHeader.AssertIncoming(); err != nil {
			log.Errorln(err)
			continue
		}

		switch msgHeader.Type {
		case 2:
			handleChatMessage(conn, payloadBuf)
		case 4:
			handleRequestLogin(conn, payloadBuf)
		case 71:
			handleRequestCreateAccount(conn, payloadBuf)
		case 105:
			handleRequestMOTD(conn, payloadBuf)
		case 107:
			handleRegObsGlobalChat(conn, payloadBuf)
		case 108:
			handleDeregObsGlobalChat(conn, payloadBuf)
		case 115:
			handleRegObsUserLogin(conn, payloadBuf)
		case 116:
			handleDeregObsUserLogin(conn, payloadBuf)
		case 168:
			handleAddGameServer(conn, payloadBuf)
		case 169:
			handleRemoveServer(conn, payloadBuf)
		case 171:
			handleRegObsServerList(conn, payloadBuf)
		case 172:
			handleDeregObsServerList(conn, payloadBuf)
		case 175:
			handleJoinServer(conn, payloadBuf)
		case 176:
			handleLeaveServer(conn, payloadBuf)
		case 177:
			handleChangeGameServer(conn, payloadBuf)
		default:
			log.Errorln("Unknown MsgType:", msgHeader.Type)
			log.Debugln(hex.EncodeToString(payloadBuf.Bytes()))
		}
	}
}

func getHeader(conn *net.TCPConn) (*packages.Header, error) {
	var header = new(packages.Header)
	var payload = make([]byte, 28)

	if _, err := conn.Read(payload); err != nil {
		return nil, err
	}

	//log.Debugln(" <-- Header:", hex.EncodeToString(payload))

	if err := binary.Read(bytes.NewBuffer(payload), binary.LittleEndian, header); err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	if err := header.AssertIncoming(); err != nil {
		return nil, fmt.Errorf("invalid header: %w", err)
	}

	log.Debugln(" <-- Header:",
		"type:", header.HeaderType, 
		"size:", header.PayloadSize,
	)

	return header, nil
}

func sendPackage(conn *net.TCPConn, hType packages.HeaderType, data []byte) error {
	retHeader := packages.NewHeader()
	retHeader.HeaderType = hType
	retHeader.PayloadSize = uint32(len(data))
	retHeader.PayloadChecksum = library.CalcChecksum(data)

	var buffer bytes.Buffer

	if err := binary.Write(&buffer, binary.LittleEndian, retHeader); err != nil {
		return fmt.Errorf("failed to create header: %v", err)
	}

	if _, err := buffer.Write(data); err != nil {
		return fmt.Errorf("failed to write payload: %v", err)
	}

	//fmt.Println(hex.EncodeToString(buffer.Bytes()))

	if _, err := buffer.WriteTo(conn); err != nil {
		return fmt.Errorf("failed to send package: %v", err)
	}

	return nil
}

func sendResult(conn *net.TCPConn, errcode uint8, errmsg string, tid uint32) {
	p := packages.NewResult(errcode, errmsg, tid)
	sendReply(conn, p, p.Type)
}

func sendReply(conn *net.TCPConn, pack any, pType uint16) {
	var buffer bytes.Buffer

	msgHeader := packages.NewMsgHeader(pType)
	if err := binary.Write(&buffer, binary.LittleEndian, msgHeader); err != nil {
		log.Errorln(err)
		return
	}

	if err := packages.Serialize(&buffer, pack); err != nil {
		log.Errorln(err)
		return
	}

	//fmt.Println(hex.EncodeToString(buffer.Bytes()))

	err := sendPackage(conn, packages.ApplicationMessage, buffer.Bytes())
	if err != nil {
		log.Errorln(err)
		return
	}

	log.Debugln(
		fmt.Sprintf(" --> %T:\n%s", pack, packages.Stringify(pack)))
}

/* NOTIFY FUNCTIONS */

func notifyUserLoggedIn(user *lobby.Account) {
	for c, a := range lobby.GetAllUsers() {
		if a.ObsUserLogin {
			p := packages.NewUserLoggedIn(user.Name, user.Uid)
			msg := fmt.Sprintf("<< %s has logged in! >>", user.Name)

			go sendReply(c, p, p.Type)
			go sendChatMessage(c, msg, 0)
		}
	}

	log.Infoln("User", user.Name, "logged in")
}

func notifyUserLoggedOut(conn *net.TCPConn) {
	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user from list")
		return
	}

	lobby.RemoveUser(conn)
	// just in case user has created a server
	lobby.RemoveServer(conn)

	for c, a := range lobby.GetAllUsers() {
		if a.ObsUserLogin {
			p := packages.NewUserLoggedOut(user.Uid)
			msg := fmt.Sprintf("<< %s has logged out >>", user.Name)

			go sendReply(c, p, p.Type)
			go sendChatMessage(c, msg, 0)
		}
	}

	log.Infoln("User", user.Name, "disconnected from server")
	conn.Close()
}

func notifyGameServerUpdate(server *lobby.Server, ticketId uint32) {
	p := createGameServerData(server, ticketId)

	for c, a := range lobby.GetAllUsers() {
		if a.ObsServerList {
			go sendReply(c, p, p.Type)
		}
	}
}

/* PACKAGE HANDLE FUNCTIONS */

func handlePackage[T any](r io.Reader) (*T, error) {
	pack := new(T)
	if err := packages.Deserialize(r, pack); err != nil {
		return nil, fmt.Errorf("failed to parse %T: %v", pack, err)
	}

	log.Debugln(
		fmt.Sprintf(" <-- %T:\n%s", pack, packages.Stringify(pack)))

	return pack, nil
}

func handleHandshake(conn *net.TCPConn) error {
	header, err := getHeader(conn)
	if err != nil {
		return err
	}
	if header.HeaderType != packages.HandshakeConnect {
		return fmt.Errorf("expected HandshakeConnect (3) but got %v", header.HeaderType)
	}
	if header.PayloadSize != 52 {
		return fmt.Errorf("expected payload size 52, but got %v", header.PayloadSize)
	}

	payload := make([]byte, header.PayloadSize)

	if _, err := conn.Read(payload); err != nil {
		return fmt.Errorf("failed to fetch handshake package: %v", err)
	}
	if header.PayloadChecksum != library.CalcChecksum(payload) {
		return fmt.Errorf("invalid checksum")
	}

	handshake := new(packages.Handshake)

	if err := binary.Read(bytes.NewBuffer(payload), binary.LittleEndian, handshake); err != nil {
		return fmt.Errorf("failed to parse handshake package: %v", err)
	}

	log.Debugln(" <-- Handshake:",
		"username:", library.CStr2Str(handshake.Username[:]),
		"password:", library.CStr2Str(handshake.Password[:]),
	)

	// HandshakeRet

	retPayload := packages.NewHandshakeRet()
	retPayload.DestID = handshake.SourceID
	retPayload.Username = handshake.Username

	var packbuf bytes.Buffer

	if err := binary.Write(&packbuf, binary.LittleEndian, retPayload); err != nil {
		return fmt.Errorf("failed to create HandshakeRet payload: %v", err)
	}

	log.Debugln(" --> HandshakeRet",
		"destID:", retPayload.DestID,
		"username:", library.CStr2Str(retPayload.Username[:]),
		"password:", hex.EncodeToString(retPayload.Password[:]),
	)

	return sendPackage(conn, packages.HandshakeConnected, packbuf.Bytes())
}

func handleRequestCreateAccount(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.RequestCreateAccount](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	/*
	* RESULT codes:
	* 0x0: OK
	* 0x1A: CD key invalid
	* 0x29: user already exists
	* 0x3E: wrong version
	*/

	if pack.Patchlevel != config.Patchlevel {
		sendResult(conn, 0x3E, "wrong patchlevel", pack.TicketId)
		return
	}
	
	// TODO CD key check (?)

	// we just accept everything, because we don't have a user database

	user := &lobby.Account{
		Name: pack.Nickname,
		Connection: conn,
	}
	lobby.AddUser(user)

	sendResult(conn, 0, "", pack.TicketId)
	notifyUserLoggedIn(user)
}

func handleRequestLogin(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.RequestLogin](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	/*
	* RESULT codes:
	* 0x00: OK
	* 0x1B: CD Key invalid
	* 0x3D: auth failed
	* 0x3E: wrong version
	*/

	if pack.Patchlevel != config.Patchlevel {
		sendResult(conn, 0x3E, "Patchlevel does not match", pack.TicketId)
		return
	}

	// TODO CD Key check (?)
	// TODO password check

	user := &lobby.Account{
		Name: pack.Nickname,
		Connection: conn,
	}
	lobby.AddUser(user)

	sendResult(conn, 0, "", pack.TicketId)
	notifyUserLoggedIn(user)
}

func handleRequestMOTD(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.RequestMOTD](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		return
	}

	p := packages.NewMOTD(config.GetMOTD(user.Name), pack.TicketId)
	sendReply(conn, p, p.Type)
}

func handleRegObsGlobalChat(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.RegObserverGlobalChat](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		return
	}
	user.ObsGlobalChat = true

	sendResult(conn, 0, "", pack.TicketId)

	log.Infoln("User", user.Name, "registered to global chat")
}

func handleRegObsServerList(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.RegObserverServerList](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		sendResult(conn, 0x3, "Cannot get user list", pack.TicketId)
		return
	}
	user.ObsServerList = true

	for _, s := range lobby.GetAllServers() {
		p := createGameServerData(s, pack.TicketId)
		sendReply(conn, p, p.Type)
	}

	sendResult(conn, 0, "", pack.TicketId)
}

func handleRegObsUserLogin(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.RegObserverUserLogin](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		return
	}
	user.ObsUserLogin = true

	sendResult(conn, 0, "", pack.TicketId)

	for _, a := range lobby.GetAllUsers() {
		if a.ObsUserLogin {
			p := packages.NewUserLoggedIn(a.Name, a.Uid)
			sendReply(conn, p, p.Type)
		}
	}

	// FIXME this also triggers when user closes server and comes back to lobby
	//msg := fmt.Sprintf("<< Welcome %s! >>", user.Name)
	//sendChatMessage(conn, msg, 0)
}

func handleDeregObsGlobalChat(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.DeregObserverGlobalChat](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		return
	}
	user.ObsGlobalChat = false

	sendResult(conn, 0, "", pack.TicketId)

	log.Infoln("User", user.Name, "de-registered from global chat")
}

func handleDeregObsUserLogin(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.DeregObserverUserLogin](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		return
	}
	user.ObsUserLogin = false

	sendResult(conn, 0, "", pack.TicketId)
}

func handleDeregObsServerList(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.DeregObserverServerList](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		return
	}
	user.ObsServerList = false

	sendResult(conn, 0, "", pack.TicketId)
}

func handleChatMessage(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.ChatMessage](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	// TODO chat commands
	// TODO chat filter

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		return
	}

	for c, a := range lobby.GetAllUsers() {
		if a.ObsGlobalChat {
			go sendChatMessage(c, pack.Txt, user.Uid)
		}
	}
}

func sendChatMessage(conn *net.TCPConn, txt string, fromId uint32) {
	p := packages.NewChat(txt, fromId)
	sendReply(conn, p, p.Type)
}

func handleAddGameServer(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.AddGameServer](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		return
	}

	log.Infoln("LOCAL ADDR:", conn.LocalAddr().String())
	log.Infoln("REMOTE ADDR:", conn.RemoteAddr().String())

	//time.Sleep(10 * time.Second)

	ip := "127.0.0.1"

	if pack.Port == 9999 { // we misuse port 9999 as error code
		log.Errorln("Client returned error code: failed to create bridge connector")
		sendResult(conn, 1, "failed to create bridge connector", pack.TicketId)
		return
	}
	if pack.Port == config.DefaultPort {
		log.Debugln("DEFAULT PORT", conn.RemoteAddr().String())
		ip = strings.Split(conn.RemoteAddr().String(), ":")[0]
	} else {
		// Public IP of Bridge Server
		// being able to have multiple bridge servers to reduce latency
		// would be great, but not worth the effort for this game
		ip = strings.Split(conn.LocalAddr().String(), ":")[0]
	}

	server := &lobby.Server{
		Name: pack.Name,
		OwnerId: user.Uid,
		Description: pack.Description,
		IP: ip,
		Port: pack.Port,
		ServerType: pack.ServerType,
		LobbyId: pack.LobbyId,
		Version: pack.Version,
		MaxPlayers: pack.MaxPlayers,
		AiPlayers: pack.AiPlayers,
		Level: pack.Level,
		GameMode: pack.GameMode,
		Hardcore: pack.Hardcore,
		Map: pack.Map,
		AutomaticJoin: pack.AutomaticJoin,
		Running: false,
		Data: pack.Data,
	}
	server.AddPlayer(conn)
	lobby.AddServer(conn, server)

	p := packages.NewResultId(0, "", server.Id, pack.TicketId)
	sendReply(conn, p, p.Type)

	notifyGameServerUpdate(server, pack.TicketId)

	log.Infoln("User", user.Name, "created a new lobby as", pack.Name)
}

func createGameServerData(server *lobby.Server, ticketId uint32) *packages.GameServerData {
	// FIXME there is an issue with server entries being listed under "other versions"

	v := server.Version // always empty (?)
	//v := "11757" // does not work
	//v := "Version 11757"
	//v := "gb_11757"
	
	p := packages.NewGameServerData()
	p.ServerId = server.Id
	p.Name = server.Name
	p.OwnerId = server.OwnerId
	p.Description = server.Description
	p.IP = server.IP
	p.Port = server.Port
	p.ServerType = server.ServerType
	p.LobbyId = server.LobbyId
	p.Version = v
	p.MaxPlayers = server.MaxPlayers
	p.CurrPlayers = uint8(server.GetPlayerCount())
	p.AiPlayers = server.AiPlayers
	p.Level = server.Level
	p.GameMode = server.GameMode
	p.Hardcore = server.Hardcore
	p.Map = server.Map
	p.Running = server.Running
	p.Data = server.Data
	p.TicketId = ticketId

	return p
}

func handleRemoveServer(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.RemoveServer](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	/*
	* TicketIds:
	* 0xB (11): RemoveGameServer
	* 0xE (14): StartGameServer 
	*/

	server, ok := lobby.GetServer(conn)

	switch tid := pack.TicketId; tid {
	case 0xB:
		if !ok {
			log.Errorln("Trying to remove server that does not exist")
			break
		}
		if server.Id != pack.ServerId {
			log.Errorln("Trying to remove server with not matching ServerID")
			break
		}

	case 0xE:
		time.Sleep(1 * time.Second) // not sure why, but this was in the original implementation
		
		if !ok {
			log.Errorln("Trying to remove server that does not exist")
			sendResult(conn, 1, "ServerID does not exit", tid)
			return
		}
		if server.Id != pack.ServerId {
			log.Errorln("Trying to remove server with not matching ServerID")
			sendResult(conn, 1, "Invalid ServerID", tid)
			return
		}
		server.Running = pack.Running

		for _, c := range server.GetPlayers() {
			p := createGameServerData(server, tid)
			go sendReply(c, p, p.Type)
		}

	default:
		log.Errorln("Unknown RemoveServer ticketID:", pack.TicketId)
	}

	lobby.RemoveServer(conn)

	for c, a := range lobby.GetAllUsers() {
		if a.ObsServerList {
			go sendReply(c, pack, pack.Type)
		}
	}

	sendResult(conn, 0, "", pack.TicketId)
}

func handleChangeGameServer(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.ChangeGameServer](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	server, ok := lobby.GetServer(conn)
	if !ok || server.Id != pack.ServerId {
		log.Errorln("ServerID problem:", server.Id, pack.ServerId)
		sendResult(conn, 3, "No server", pack.TicketId)
		return
	}

	server.Name = pack.Name
	server.Description = pack.Description
	//server.MaxPlayers = pack.MaxPlayers - pack.SlotsOccupied
	//server.AiPlayers = 
	server.MaxPlayers = pack.MaxPlayers
	server.AiPlayers = pack.SlotsOccupied
	server.Level = pack.Level
	server.GameMode = pack.GameMode
	server.Hardcore = pack.Hardcore
	server.Map = pack.Map
	server.Running = pack.Running
	server.Data = pack.Data
	server.PropertyMask = pack.PropertyMask

	notifyGameServerUpdate(server, pack.TicketId)
	sendResult(conn, 0, "", pack.TicketId)

	log.Infoln("Server", server.Name, "got updated")
}

func handleJoinServer(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.JoinServer](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	/*
	* Error codes:
	* 0x84 (132): GameServer not found
	* 0x87 (135): GameServer full
	*/

	server, ok := lobby.GetServerById(pack.ServerId)
	if !ok {
		log.Errorln("Tried to join ServerId that does not exist")
		sendResult(conn, 0x84, "game server not found", pack.TicketId)
		return
	}

	if server.IsFull() {
		log.Infoln("Lobby", server.Name, "is already full")
		sendResult(conn, 0x87, "game server full", pack.TicketId)
		return
	}

	server.AddPlayer(conn)

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		sendResult(conn, 0x84, "game server not found", pack.TicketId)
		return
	}

	user.JoinedServer = server

	sendResult(conn, 0, "", pack.TicketId)
}

func handleLeaveServer(conn *net.TCPConn, r io.Reader) {
	pack, err := handlePackage[packages.LeaveServer](r)
	if err != nil {
		log.Errorln(err)
		return
	}

	user, ok := lobby.GetUser(conn)
	if !ok {
		log.Errorln("Failed to fetch user")
		return
	}

	if user.JoinedServer != nil {
		user.JoinedServer.RemovePlayer(conn)
		user.JoinedServer = nil

		sendResult(conn, 0, "", pack.TicketId)
	} else {
		sendResult(conn, 1, "user has not joined any server", pack.TicketId)
	}
}
