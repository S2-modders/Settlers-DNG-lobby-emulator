package netbridge

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"s2dnglobby/config"
	"s2dnglobby/library"
	"strconv"
	"strings"
	"sync"
	"time"
)

var log = library.GetLogger("BridgeController")

const portStart = 10_000
const portRange = 1000

var portLock sync.Mutex

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

func getNextAvailablePort(start int) (int, error) {
	for i := 0; i < portRange; i++ {
		p := start + i
		b, err := requestAvailablePort(p)
		if err != nil {
			return 0, err
		}

		if b {
			return p, nil
		}
	}
	return 0, fmt.Errorf("no available port found")
}

func checkPortForward(ip string, port int) bool {
	newAddr := fmt.Sprintf("%s:%d", ip, config.DefaultPort)

	conn, err := net.DialTimeout("tcp", newAddr, 2 * time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}


func handleForwardCheck(w http.ResponseWriter, r *http.Request) {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	log.Debugln("Checking port forward for", ip)

	if checkPortForward(ip, config.DefaultPort) {
		log.Debugln("Direct connect possible")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, ip)
	} else {
		log.Debugln("Direct connect failed")
		w.WriteHeader(900)
	}
}

func handleControllerPort(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, config.CONTROLLER_PORT)
}

func handleBridgePort(w http.ResponseWriter, r *http.Request) {
	// TODO maybe do IP check (?)

	portLock.Lock()
	defer portLock.Unlock()

	port, err := getNextAvailablePort(portStart)
	if err != nil {
		log.Errorln("failed to fetch available port:", err)

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, port)

	log.Infoln("host port requested; found:", port)
}

func InitBridgeController() {
	// check if port forward is working
	http.HandleFunc("/port/check", handleForwardCheck)

	// request API port (controller port used by FRP)
	http.HandleFunc("/port/controller", handleControllerPort)

	// request public port for bridge connection
	http.HandleFunc("/request/port", handleBridgePort)

	go http.ListenAndServe(fmt.Sprintf(":%d", config.API_PORT), nil)

	log.Infoln("API listening on port", config.API_PORT)
}
