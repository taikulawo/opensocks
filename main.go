package main

import (
	"flag"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"strconv"
)

var(
	isClient = flag.Bool("isClient",true,"True for client mode")
	p = flag.Int("p",9999,"listen port")
)

func main(){
	flag.Parse()
	filehook := filename.NewHook()
	logrus.SetOutput(os.Stdout)
	logrus.AddHook(filehook)
	addr, err := net.ResolveTCPAddr("tcp","0.0.0.0:" + strconv.Itoa(*p))
	if err != nil{
		panic(err)
	}

	logrus.Infof("Start Listen on %d",addr.Port)
	ln, err := net.ListenTCP("tcp4",addr)

	if err != nil{
		panic(err)
	}
	for{
		conn, err := ln.AcceptTCP()
		if err != nil{
			conn.Close()
			return
		}
		logrus.Infof("Recv Conn from %s",conn.RemoteAddr())
		var handler InboundHandler
		var s int32
		if *isClient{
			s = StageInit
		}else{
			s = StageRunning
		}
		handler = &SocksHandler{
			stage: s,
			isClient:*isClient,
		}
		go handler.Handle(conn)
	}

}