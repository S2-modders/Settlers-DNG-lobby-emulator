package netbridge

import (
	"os"
	"os/exec"
	"s2dnglobby/config"
	"strconv"
)

/*
func runBridge(hostConn *net.TCPConn, clientSocket *net.TCPListener) {
	//defer clientSocket.Close()

	// we only accept one client for now

	conn, err := clientSocket.AcceptTCP()
	if err != nil {
		runBridge(hostConn, clientSocket)
	}
	conn.SetNoDelay(true)
	conn.SetReadBuffer(4096)
	conn.SetWriteBuffer(4096)

	go func() {
		buf := make([]byte, 4096)
		_, err := io.CopyBuffer(hostConn, conn, buf)
		if err != nil {
			clientSocket.Close()
			conn.Close()
			hostConn.Close()
		}
	}()

	
	go func() {
		buf := make([]byte, 4096)
		_, err := io.CopyBuffer(conn, hostConn, buf)
		if err != nil {
			clientSocket.Close()
			conn.Close()
			hostConn.Close()
		}
	}()
}

func createBridge(hostPort, clientPort int) error {
	log.Infoln("Creating bridge between", hostPort, "and", clientPort)
	log.Debugln("Waiting for host to connect on", hostPort)

	hAddr := net.TCPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: hostPort,
	}
	hListener, err := net.ListenTCP("tcp", &hAddr)
	if err != nil {
		return err
	}
	
	hListener.SetDeadline(time.Now().Add(5 * time.Second))
	hConn, err := hListener.AcceptTCP()
	hListener.Close()
	if err != nil {
		return err
	}
	hConn.SetNoDelay(true)
	hConn.SetReadBuffer(4096)
	hConn.SetWriteBuffer(4096)

	log.Debugln("Host", hConn.RemoteAddr(), "connected")

	log.Debugln("Opening client port", clientPort)

	cAddr := net.TCPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: clientPort,
	}
	cListener, err := net.ListenTCP("tcp", &cAddr)
	if err != nil {
		return err
	}

	go runBridge(hConn, cListener)
	return nil
}
*/

func runBridgeConnector() {
	cmd :=  exec.Command("./frps", "-p", strconv.Itoa(config.CONTROLLER_PORT))
	cmd.Stdout = os.Stdout
	cmd.Run()
}
