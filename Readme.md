## The Settlers II: 10th Lobby Emulator [WiP]

This project is an attempt to recreate the online mode of The Settlers II: 10th anniversary edition by emulating the online lobby and reimplementing the tincat3 network protocol.

Tincat version used: 3.0.53

### Current Progress:

- [x] create account
- [x] login with account
- [x] request and show MOTD
- [x] show online status of other players
- [x] global chat with properly working usernames
- [x] error messages when auth or account creation failed
- [ ] create new game
- [ ] see all created games
- [ ] join new game
- [ ] launch new lobby with other players

### Usage [outdated - TODO]

In the directory of the executable write the public IP of the lobby server into ip.cfg file.

1. run game
2. go multi player
3. create account
4. log in
5. make nickname
6. make charactor
7. enter lobby
8. run server, example in starts2gs.cmd file
9. refresh server list in lobby
10. join game
11. play

The server hosting needs an account, but may use the same credentials like the player accounts.

### Credits

- BIG THANKS to cocomed who originally created this project and wrote most of the core code [here](http://darkmatters.org/forums/index.php?/topic/23833-network-traffic-probes-for-sacred-2-available/&do=findComment&comment=7015188)
- pnxr for continuing the project and adding fixes
- the sacred2 community
