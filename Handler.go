package main

import "net"

type InboundHandler interface{
	Handle(conn *net.TCPConn)
}



