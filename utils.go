package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"net"
)

type PipeHandler func([]byte)

func LimitReader (conn *net.TCPConn, count int64)(int,error,[]byte){
	reader := io.LimitReader(conn,count)
	bs := make([]byte,count)
	c,err := reader.Read(bs)
	return c,err, bs
}

func PipeStart(reader io.Reader, writer io.Writer, ps ...PipeHandler)error{
	for{
		bs := make([]byte,256)
		count, err := reader.Read(bs)
		bs = bs[:count]
		if err != nil{
			logrus.Errorf("Connection Closed %v",err)
			return err
		}
		for _, h := range ps{
			h(bs)
		}
		if _, err := writer.Write(bs); err != nil{
			return err
		}
	}
}