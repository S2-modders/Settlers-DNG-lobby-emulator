## The Settlers II: 10th Lobby Emulator [WiP]

This project is an attempt to recreate the online mode of The Settlers II: 10th anniversary edition by emulating the online lobby and reimplementing the tincat3 network protocol.

Tincat version used: 3.0.53

### Current Progress:

- [x] create account (just a stub, no actual account creation happening)
- [x] login with account
- [x] request and show MOTD
- [x] show online status of other players
- [x] global chat with properly working usernames
- [x] error messages when auth or account creation failed
- [x] create new game
- [x] join new game
- [x] launch new lobby with other players
- [x] port check when hosting game, prefer direct connection
- [x] automatic creation of TCP bridge if direct connection fails
- [ ] automatic disconnect from TCP bridge when user leaves multiplayer screen
- [ ] see all created games with default filter (cannot get this to work :(( - kinda workaround with dll hack for now)

### Credits

- BIG THANKS to cocomed who originally created the C# implementation this port is based on [here](http://darkmatters.org/forums/index.php?/topic/23833-network-traffic-probes-for-sacred-2-available/&do=findComment&comment=7015188)
- pnxr for continuing the project and adding fixes
- the sacred2 community
