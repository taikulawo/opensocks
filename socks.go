package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
)

const(
	StageInit = iota
	StageConnecting
	StageRunning
)

const(
	Version = 0x05

	CMD_CONNECT = 0x01

	ATYP_V4 = 0x01
	ATYP_DOMAIN_NAME = 0x03

)

const(
	InitHeaderLen = 3
	RequestHandlerLen = 3
)

type NetAddress struct{
	Host string
	Port int
}

type SocksHandler struct{
	stage int32
	isClient bool
}

func (h *SocksHandler) Handle(conn *net.TCPConn){
	if !h.isClient{
		ps := h.StartCryptoStream([]byte(AuthPassword))
		_, err, res := LimitReader(conn,2)

		ps(res)

		dstLen := int(res[1]) + 2
		_, err, dstPart := LimitReader(conn,int64(dstLen))
		ps(dstPart)

		res = append(res,dstPart...)

		dstAddr, err  := h.handleCmdConnect(res)

		if err != nil{
			logrus.Errorf(err.Error())
			return
		}

		remoteConn, err := h.connectToRemote(*dstAddr)

		if err != nil{
			logrus.Errorf(err.Error())
			return
		}
		go func(){
			ps := h.StartCryptoStream([]byte(AuthPassword))
			PipeStart(remoteConn,conn,ps)
		}()
		PipeStart(conn,remoteConn,ps)

	}
	_,err,initHandler := LimitReader(conn,InitHeaderLen)
	if err != nil{
		logrus.Debugf("Error when read socks init header >> %v",err)
		return
	}
	if initHandler[0] != 0x05{
		logrus.Debugf("Only support Socks Version 5 , >> current %d",initHandler[0])
		conn.Close()
		return
	}
	_, err = conn.Write([]byte{0x05,0x00})
	if err != nil{
		logrus.Debugf("Error when write init header to local, >> %v",err)
		conn.Close()
		return
	}
	_, err, bs := LimitReader(conn,RequestHandlerLen)
	if bs[1] == CMD_CONNECT{
		conn.Write([]byte{0x05,0x00,0x00,0x01,0x00,0x00,0x00,0x00,0x00,0x00})
		remoteConn ,err := net.DialTCP("tcp",nil,&net.TCPAddr{
			Port:ServerPort,
			IP:net.ParseIP(ServerIp),
		})

		if err != nil{
			logrus.Errorf("Cannot connect to remote >> %v",err)
			return
		}

		go func(){
			ps := h.StartCryptoStream([]byte(AuthPassword))
			PipeStart(remoteConn,conn,ps)
		}()

		processStream := h.StartCryptoStream([]byte(AuthPassword))
		PipeStart(conn,remoteConn,processStream)
	}
}

func (h *SocksHandler)handleCmdConnect(bs []byte)(*NetAddress,error){
	switch bs[0]{
	case ATYP_DOMAIN_NAME:{
		nameLen := bs[1]
		domainName := string(bs[2:nameLen + 2])
		rawPort := bs[len(bs) - 2:]
		port := binary.BigEndian.Uint16(rawPort)
		return &NetAddress{
			Host:domainName,
			Port:int(port),
		},nil
	}
	default:{
		return nil, errors.New(fmt.Sprintf("Not support ATYP >> current %d",bs[0]))
	}
	}
}

func (h *SocksHandler)connectToRemote(addr NetAddress)(net.Conn, error){

	conn, err := net.Dial("tcp",addr.Host + ":" + strconv.Itoa(addr.Port))
	if err != nil{
		return nil, err
	}
	return conn,nil
}

func (h *SocksHandler)StartCryptoStream(key []byte)PipeHandler{
	cipher := NewRC4(key)
	return func(bs []byte){
		cipher.XORKeyStream(bs,bs)
	}
}

