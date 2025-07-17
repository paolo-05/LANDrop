package utils

import (
	"net"
)

func GetLocalIP() string {
	// Connect to an external address to get the interface used for outbound traffic
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
