package packages

import (
	"fmt"
)

//{0xEF, 0xFB, 0xBA, 0xDA}

type HeaderType uint32

const (
	ApplicationMessage HeaderType = 2
	HandshakeConnect   HeaderType = 3
	HandshakeConnected HeaderType = 5
	Ping               HeaderType = 11
)

const header_magic = 0xDABAFBEF // EFFBBADA
const server_ID = 0xEFFFFFCC    // CCFFFFEF
const client_ID = 0xEFFFFFEE    // EEFFFFEF

type Header struct {
	Magic           uint32
	SourceID        uint32
	DestID          uint32
	HeaderType      HeaderType
	Unknown         uint32
	PayloadSize     uint32
	PayloadChecksum uint32
}

func (h *Header) AssertIncoming() error {
	if h.Magic != header_magic {
		return fmt.Errorf("invalid header magic: %d", h.Magic)
	}
	if h.SourceID != client_ID {
		return fmt.Errorf("invalid client ID: %d", h.SourceID)
	}
	if h.DestID != server_ID {
		return fmt.Errorf("invalid server ID: %d", h.DestID)
	}
	if h.PayloadSize > 1024 {
		return fmt.Errorf("payload size > 1024 is VERY SUS: %d", h.PayloadSize)
	}

	return nil
}
func NewHeader() *Header {
	return &Header{
		Magic:    header_magic,
		SourceID: server_ID,
		DestID:   3,
		Unknown:  0,
	}
}

type Handshake struct {
	Magic    uint32
	SourceID uint32
	Username [32]byte
	Password [8]byte
	Unknown  uint32
}

type HandshakeRet struct {
	Magic    uint32
	DestID   uint32
	Username [32]byte
	Password [8]byte
	Unknown  uint32
}

func NewHandshakeRet() *HandshakeRet {
	return &HandshakeRet{
		Magic:    header_magic,
		Unknown:  0,
		Password: [8]byte{0x2D, 0, 0, 0, 0, 0, 0, 0},
	}
}

const payload_magic = 0x27D8 // D827

type MsgHeader struct {
	Magic uint16
	Type  uint16
}

func (h *MsgHeader) AssertIncoming() error {
	if h.Magic != payload_magic {
		return fmt.Errorf("invalid payload header magic: %d", h.Magic)
	}

	return nil
}
func NewMsgHeader(msgType uint16) MsgHeader {
	return MsgHeader{
		Magic: payload_magic,
		Type:  msgType,
	}
}

/* MSG DEFS */

// 2
type ChatMessage struct {
	Type     uint16
	Mode     uint32
	Txt      string
	TicketId uint32
	//FromId uint32 // package contains some garbage data here
}

// 4
type RequestLogin struct {
	Type       uint16
	Nickname   string
	Password   string
	Cdkey      []byte
	Keypool    uint16
	Patchlevel uint32
	TicketId   uint32
}

// 42
type Result struct {
	Type      uint16
	ErrorCode uint8
	ErrorMsg  string
	TicketId  uint32
}

func NewResult(errorCode uint8, msg string, ticketId uint32) *Result {
	return &Result{
		Type:      42,
		ErrorCode: errorCode,
		ErrorMsg:  msg,
		TicketId:  ticketId,
	}
}

// 71
type RequestCreateAccount struct {
	Type       uint16
	Nickname   string
	Password   string
	Cdkey      []byte
	Keypool    uint16
	Patchlevel uint32
	TicketId   uint32
}

// 105
type RequestMOTD struct {
	Type     uint16
	TicketId uint32
}

// 106
type MOTD struct {
	Type     uint16
	Txt      string
	TicketId uint32
}

func NewMOTD(txt string, ticketId uint32) *MOTD {
	return &MOTD{
		Type:     106,
		Txt:      txt,
		TicketId: ticketId,
	}
}

// 107
type RegObserverGlobalChat struct {
	Type     uint16
	TicketId uint32
}

// 108
type DeregObserverGlobalChat struct {
	Type     uint16
	TicketId uint32
}

// 109
type UserLoggedIn struct {
	Type   uint16
	UserId uint32
	Name   string
}

func NewUserLoggedIn(name string, userId uint32) *UserLoggedIn {
	return &UserLoggedIn{
		Type:   109,
		Name:   name,
		UserId: userId,
	}
}

// 110
type UserLoggedOut struct {
	Type   uint16
	UserId uint32
}

func NewUserLoggedOut(userId uint32) *UserLoggedOut {
	return &UserLoggedOut{
		Type:   110,
		UserId: userId,
	}
}

// 115
type RegObserverUserLogin struct {
	Type     uint16
	SendAll  bool
	TicketId uint32
}

// 116
type DeregObserverUserLogin struct {
	Type     uint16
	TicketId uint32
}

// 153
type ResultId struct {
	Type      uint16
	ErrorCode uint8
	ErrorMsg  string
	Id        uint32
	TicketId  uint32
}

func NewResultId(code uint8, msg string, id, ticketId uint32) *ResultId {
	return &ResultId{
		Type:      153,
		ErrorCode: code,
		ErrorMsg:  msg,
		Id:        id,
		TicketId:  ticketId,
	}
}

// 165
type Chat struct {
	Type   uint16
	Txt    string
	FromId uint32
}

func NewChat(txt string, fromId uint32) *Chat {
	return &Chat{
		Type:   165,
		Txt:    txt,
		FromId: fromId,
	}
}

// 168
type AddGameServer struct {
	Type          uint16
	Name          string
	Description   string
	Port          uint32
	ServerType    uint8
	LobbyId       uint32
	Version       string
	MaxPlayers    uint8
	AiPlayers     uint8
	Level         uint8
	GameMode      uint8
	Hardcore      bool
	Map           string
	AutomaticJoin bool
	Data          []byte
	TicketId      uint32
}

// 169
type RemoveServer struct {
	Type     uint16
	ServerId uint32
	Running  bool
	TicketId uint32
}

func NewRemoveServer(serverId uint32, running bool, ticketId uint32) *RemoveServer {
	return &RemoveServer{
		Type:     169,
		ServerId: serverId,
		Running:  running,
		TicketId: ticketId,
	}
}

// 170
type GameServerData struct {
	Type        uint16
	ServerId    uint32
	Name        string
	OwnerId     uint32
	Description string
	IP          string
	Port        uint32
	ServerType  uint8
	LobbyId     uint32
	Version     string
	MaxPlayers  uint8
	CurrPlayers uint8
	AiPlayers   uint8
	Level       uint8
	GameMode    uint8
	Hardcore    bool
	Map         string
	Running     bool
	Data        []byte
	TicketId    uint32
}

func NewGameServerData() *GameServerData {
	return &GameServerData{
		Type: 170,
	}
}

// 171
type RegObserverServerList struct {
	Type       uint16
	SendAll    bool
	ServerType uint8
	RoomId     uint32
	Selection  uint32
	TicketId   uint32
}

// 172
type DeregObserverServerList struct {
	Type     uint16
	TicketId uint32
}

// 175
type JoinServer struct {
	Type     uint16
	Unused   uint32 // UserId field is not used, always 0
	ServerId uint32
	TicketId uint32
}

// 176
type LeaveServer struct {
	Type     uint16
	Unused   uint32 // UserId field is not used, always 0
	TicketId uint32
}

// 177
type ChangeGameServer struct {
	Type          uint16
	ServerId      uint32
	Name          string
	Description   string
	MaxPlayers    uint8
	SlotsOccupied uint8 // Blocked slots either by AI or by closing it # AiPlayers
	Level         uint8
	GameMode      uint8
	Hardcore      bool
	Map           string
	Running       bool
	Data          []byte
	PropertyMask  uint32
	TicketId      uint32
}
