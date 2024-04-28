package tracker

import (
	"bittorrent-client/logger"
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/url"
)

// "Magic" const as per https://www.bittorrent.org/beps/bep_0015.html
const magicProtocolID uint64 = 0x41727101980

const (
	actionConnect  uint32 = 0
	actionAnnounce uint32 = 2
)

type ConnectRequest struct {
	ProtocolID    uint64
	Action        uint32
	TransactionID uint32
}

func (r *ConnectRequest) WriteTo(w io.Writer) (int64, error) {
	return 0, binary.Write(w, binary.BigEndian, r)
}

func createConnectRequest() *ConnectRequest {
	return &ConnectRequest{
		ProtocolID:    magicProtocolID,
		Action:        actionConnect,
		TransactionID: uint32(rand.Int31()),
	}
}

type ConnectResponse struct {
	Action        uint32
	TransactionId uint32
	ConnectionId  uint64
}

func parseConnectResponse(r []byte) *ConnectResponse {
	return &ConnectResponse{
		Action:        binary.BigEndian.Uint32(r[0:4]),
		TransactionId: binary.BigEndian.Uint32(r[4:8]),
		ConnectionId:  binary.BigEndian.Uint64(r[8:]),
	}
}

type UDPTracker struct {
	rawURL      string
	destination string
	log         logger.Logger
}

func New(rawURL string) *UDPTracker {
	parsedURL, _ := url.Parse(rawURL)
	return &UDPTracker{
		rawURL:      rawURL,
		destination: parsedURL.Host,
		log:         logger.New(),
	}
}

func (t *UDPTracker) Connect() *ConnectResponse {
	udpAddr, err := net.ResolveUDPAddr("udp4", t.destination)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return nil
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		t.log.Error().Msgf("%s%v", "Error creating UDP connection:", err)
		return nil
	}
	defer conn.Close()

	req := createConnectRequest()
	_, err = req.WriteTo(conn)

	if err != nil {
		t.log.Error().Msgf("%s%v", "Error sending connect request:", err)
		return nil
	}

	resp := make([]byte, 16)
	_, err = bufio.NewReader(conn).Read(resp)

	if err != nil {
		t.log.Error().Msgf("%s%v", "Error reading connect response:", err)
		return nil
	}

	return parseConnectResponse(resp)
}
