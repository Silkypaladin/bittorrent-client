package main

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"net/url"
)

// "Magic" const as per https://www.bittorrent.org/beps/bep_0015.html
const PROTOCOL_ID uint64 = 0x41727101980

const (
	ACTION_CONNECT  uint32 = 0
	ACTION_ANNOUNCE uint32 = 2
)

type ConnectResponse struct {
	Action        uint32
	TransactionId uint32
	ConnectionId  uint64
}

func getPeersList(announce string) *ConnectResponse {
	addr, _ := url.Parse(announce)
	udpAddr, err := net.ResolveUDPAddr("udp4", addr.Host)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return nil
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		fmt.Println("Error creating UDP connection:", err)
		return nil
	}
	defer conn.Close()

	req := createConnectRequest()
	_, err = conn.Write(req)

	if err != nil {
		fmt.Println("Error sending connect request:", err)
		return nil
	}

	resp := make([]byte, 16)
	_, err = bufio.NewReader(conn).Read(resp)

	if err != nil {
		fmt.Println("Error reading connect response:", err)
		return nil
	}

	return parseConnectResponse(resp)
}

func createConnectRequest() []byte {
	b := make([]byte, 16)

	binary.BigEndian.PutUint64(b[0:8], PROTOCOL_ID)
	binary.BigEndian.PutUint32(b[8:], ACTION_CONNECT)

	// Assign random transaction id, https://www.bittorrent.org/beps/bep_0015.html
	rand.Read(b[12:])

	return b
}

func parseConnectResponse(r []byte) *ConnectResponse {
	return &ConnectResponse{
		Action:        binary.BigEndian.Uint32(r[0:4]),
		TransactionId: binary.BigEndian.Uint32(r[4:8]),
		ConnectionId:  binary.BigEndian.Uint64(r[8:]),
	}
}
