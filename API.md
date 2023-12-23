# Tincat3 protocol used by S2 DNG

## Package Flow

```mermaid
sequenceDiagram
participant C as Client
participant S as Server
participant CN as Clients n

note over C, CN: Handshake

C ->> S: HandshakeConnect
S -->> C: HandshakeConnected

note over C, CN: Register

C ->> S: RequestCreateAccount (71)
S -->> C: Result (42)
S ->> CN: UserLoggedIn (109)

note over C, CN: Login

C ->> S: RequestLogin (4)
S -->> C: Result (42)
S ->> CN: UserLoggedIn (109)

C ->> S: RequestMOTD (105)
S -->> C: MOTD (106)

C ->> S: RegObserverGlobalChat (107)
S -->> C: Result (42)

C ->> S: RegObserverServerList (171)
S -->> C: Result (42)
loop for all available servers
    S -->> C: GameServerData (170)
end

C ->> S: RegObserverUserLogin (115)
S -->> C: Result (42)
loop for all logged in users
    S -->> C: UserLoggedIn (109)
end

note over C, CN: Logout

C ->> S: DeregObserverUserLogin (116)
S -->> C: Result (42)

C ->> S: DeregObserverServerList (172)
S -->> C: Result (42)

C ->> S: DeregObserverGlobalChat (108)
S -->> C: Result (42)

note over C, CN: on client disconnect
S -X C: Disconnect
S ->> CN: UserLoggedOut (110)

note over C, CN: Lobby

C ->> S: ChatMessage (2)
S ->> CN: Chat (165)

C ->> S: AddGameServer (168)
S -->> C: ResultId (153)
S ->> CN: GameServerData (170)

C ->> S: RemoveServer / Remove (169 / 11)
S ->> CN: RemoveServer (169)

C ->> S: RemoveServer / Start (169 / 14)
S ->> CN: GameServerData (170)

C ->> S: ChangeGameServer (177)
S ->> CN: GameServerData (170)
S -->> C: Result (42)

C ->> S: JoinServer (175)
S -->> C: Result (42)

C ->> S: LeaveServer (176)
S -->> C: Result (42)
```

## Headers & IDs

magic: `EFFBBADA` \
payloadMagic: `D827` \
serverID: `CCFFFFEF`

### Header Types

2: ApplictaionMessage \
3: HandshakeConnect \
5: HandshakeConnected \
11: Ping

### Payloads

see `packages.go`
