package config

import (
	"fmt"
)

const DEBUGGING = true

const DefaultPort = 5479
const SERVER_PORT = 6800 // port of the lobby server
const API_PORT = 6801 // port of the HTTP API of lobby server
const CONTROLLER_PORT = 6802 // port of FRP controller

const Patchlevel = 11757

const VersionMaj = 0;
const VersionMin = 2;
const Year = "2022 - 2023"

const MOTD = `Welcome to The Settlers II: 10th anniversary! 
--- you are logged in as %s --- 

S2 online lobby by zocker_160, cocomed and pnxr
v%d.%d-alpha %s

Join our Discord: https://discord.gg/UAXH3VS9Qy`

const ConfigFileName = ""

func GetMOTD(name string) string {
	return fmt.Sprintf(
		MOTD, name, VersionMaj, VersionMin, Year)
}
