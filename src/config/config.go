package config

import (
	"fmt"
	"os"
)

const DEBUGGING = true

const DefaultPort = 5479
const SERVER_PORT = 6800 // port of the lobby server
const API_PORT = 6801 // port of the HTTP API of lobby server

const PORT_START = 10_000
const PORT_END   = 11_000

const CONTROLLER_PORT = 2222
const CONTROLLER_USERNAME = "siedler"
const CONTROLLER_PASSWORD = "ilovesettlers"

const Patchlevel = 11757
//const Patchlevel = 9212

const VersionMaj = 0;
const VersionMin = 3;
const Year = "2022 - 2025"

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


type ControllerData struct {
	Port int
	Username string
	Password string
}


func GetControllerData() *ControllerData {
	username, ok := os.LookupEnv("CONTROLLER_USERNAME")
	if !ok {
		username = CONTROLLER_USERNAME
	}

	password, ok := os.LookupEnv("CONTROLLER_PASSWORD")
	if !ok {
		password = CONTROLLER_PASSWORD
	}

	return &ControllerData{
		Port: CONTROLLER_PORT,
		Username: username,
		Password: password,
	}
}
