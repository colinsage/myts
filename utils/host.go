package utils

import (
	"net"
	"strconv"
)

func GetAdminHost(tcpHost string) string {
	//tcp port add 1
	host, port, _ := net.SplitHostPort(tcpHost)

	p,_ := strconv.Atoi(port)
	p -= 2

	port = strconv.Itoa(p)
	return net.JoinHostPort(host, port)
}
