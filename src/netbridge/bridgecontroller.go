package netbridge

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"s2dnglobby/config"
	"s2dnglobby/library"
	"strconv"
	"sync"
	"time"
)

var log = library.GetLogger("BridgeController")


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

func getNextAvailablePort() (int, error) {
	for i := config.PORT_START; i < config.PORT_END; i++ {
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

func checkPortForward(ip string, port int) bool {
	addr := net.JoinHostPort(ip, strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", addr, 2 * time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}


func handleForwardCheck(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Errorln("Failed to parse RemoteAddr")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Debugln("Checking port forward for", host)

	if checkPortForward(host, config.DefaultPort) {
		log.Debugln("Direct connect possible")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, host)
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

	port, err := getNextAvailablePort()
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

func handleCredReq(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	data := struct {
		Port int
		Username string
		Password string
	}{
		Port: config.CONTROLLER_PORT,
		Username: config.CONTROLLER_USERNAME,
		Password: config.CONTROLLER_PASSWORD,
	}
	encoder.Encode(data)
}


func InitBridgeController() {
	// check if port forward is working
	http.HandleFunc("/port/check", handleForwardCheck)

	// request public port for bridge connection
	http.HandleFunc("/port/request", handleBridgePort)

	// request SSH port for bridge connector
	//http.HandleFunc("/port/controller", handleControllerPort)
	
	// request SSH credentials for login
	http.HandleFunc("/bridge/get", handleCredReq)


	go http.ListenAndServe(fmt.Sprintf(":%d", config.API_PORT), nil)
	log.Infoln("API listening on port", config.API_PORT)

	// maybe TODO custom bridge server
}
