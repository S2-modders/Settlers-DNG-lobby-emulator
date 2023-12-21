package netbridge

import (
	"fmt"
	"os/exec"
	"s2dnglobby/library"
	"strconv"
)

var log = library.GetLogger("BridgeController")

const basePort = 10000
const maxPort = 65000

func requestAvailablePort(port int) (bool, error) {
	ss := exec.Command("ss", "-tulpn")
	grep := exec.Command("grep", strconv.Itoa(port))
	wc := exec.Command("wc", "-l")

	ssPipe, _ := ss.StdoutPipe()
	defer ssPipe.Close()
	grep.Stdin = ssPipe

	grepPipe, _ := grep.StdoutPipe()
	defer grepPipe.Close()
	wc.Stdin = grepPipe

	ss.Start()
	grep.Start()
	
	res, err := wc.Output()
	if err != nil {
		return false, err
	}

	if len(res) < 1 {
		return false, fmt.Errorf("no output")
	}

	i, err := strconv.Atoi(string(res[0]))
	return i == 0, err
}

func getNextAvailablePort() (int, error) {
	for i := basePort; i < maxPort; i++ {
		b, err := requestAvailablePort(i)
		if err != nil {
			return 0, err
		}

		if b {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no available port found")
}

func InitBridgeController() {
	
	p, err := getNextAvailablePort()
	if err != nil {
		log.Errorln(err)
	} else {
		log.Debugln("Port", p, "available")
	}

	// TODO small HTTP REST API server with /api/request endpoint

	log.Infoln("ready")
}
