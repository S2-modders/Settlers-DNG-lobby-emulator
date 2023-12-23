package lobby

import (
	"net"
	"s2dnglobby/library"
	"sync"
	"sync/atomic"
	"time"
)

var log = library.GetLogger("Lobby")

type Account struct {
	Name string
	//Password string
	//Cdkey []byte
	//Keypool int
	//Patchlevel int

	Connection *net.TCPConn
	Uid uint32

	ObsUserLogin bool
	ObsGlobalChat bool
	ObsServerList bool

	JoinedServer *Server
}

var users = make(map[*net.TCPConn]*Account)
var userIdCounter atomic.Uint32
var usersLock sync.RWMutex

func AddUser(user *Account) {
	id := userIdCounter.Add(1)
	user.Uid = id

	usersLock.Lock()
	users[user.Connection] = user
	usersLock.Unlock()
}

func RemoveUser(conn *net.TCPConn) {
	usersLock.Lock()
	delete(users, conn)
	usersLock.Unlock()
}

func GetUser(conn *net.TCPConn) (*Account, bool) {
	usersLock.RLock()
	val, ok := users[conn]
	usersLock.RUnlock()

	return val, ok
}

func GetAllUsers() map[*net.TCPConn]*Account {
	return users
}

type Server struct {
	Name string
	OwnerId uint32
	Description string
	IP string
	Port uint32
	ServerType uint8
	LobbyId uint32
	Version string
	MaxPlayers uint8
	AiPlayers uint8
	Level uint8
	GameMode uint8
	Hardcore bool
	Map string
	AutomaticJoin bool
	Running bool
	Data []byte

	PropertyMask uint32 // not sure what this is for

	Id uint32
	players []*net.TCPConn
	playersLock sync.Mutex
}
func (s *Server) AddPlayer(conn *net.TCPConn) {
	s.playersLock.Lock()
	s.players = append(s.players, conn)
	s.playersLock.Unlock()
}
func (s *Server) RemovePlayer(conn *net.TCPConn) {
	var newList []*net.TCPConn

	for _, c := range s.players {
		if c == conn {
			continue
		}
		newList = append(newList, c)
	}

	s.playersLock.Lock()
	s.players = newList
	s.playersLock.Unlock()
}
func (s *Server) GetPlayerCount() int {
	return len(s.players)
}
func (s *Server) GetPlayers() []*net.TCPConn {
	return s.players
}
func (s *Server) IsFull() bool {
	return s.GetPlayerCount() + int(s.AiPlayers) >= int(s.MaxPlayers)
}


var servers = make(map[*net.TCPConn]*Server)
var serverIdCounter atomic.Uint32
var serversLock sync.RWMutex

func AddServer(conn *net.TCPConn, server *Server) {
	id := serverIdCounter.Add(1)
	server.Id = id

	serversLock.Lock()
	servers[conn] = server
	serversLock.Unlock()
}

func RemoveServer(conn *net.TCPConn) {
	serversLock.Lock()
	delete(servers, conn)
	serversLock.Unlock()
}

func GetServer(conn *net.TCPConn) (*Server, bool) {
	serversLock.RLock()
	val, ok := servers[conn]
	serversLock.RUnlock()

	return val, ok
}

func GetServerById(serverId uint32) (*Server, bool) {
	serversLock.RLock()
	defer serversLock.RUnlock()

	for _, s := range servers {
		if s.Id == serverId {
			return s, true
		}
	}
	return nil, false
}

func GetAllServers() map[*net.TCPConn]*Server {
	return servers
}

/* Lobby main loops */

func InitLobby() {
	go statsPrinter()

	log.Infoln("Lobby initialized")
}

func statsPrinter() {
	for {
		time.Sleep(10 * time.Second)
		log.Infoln(
			"Connected users:", len(users),
			"Last UID:", userIdCounter.Load(),
			"Created rooms:", len(servers),
		)
	}
}
