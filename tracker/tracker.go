package tracker

import (
	"bittorrent-client/logger"
	"bittorrent-client/torrent"
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
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
	actionAnnounce uint32 = 1
)

type Event uint32

const (
	EventNone Event = iota
	EventCompleted
	EventStarted
	EventStopped
)

type messageHeader struct {
	Action        uint32
	TransactionID uint32
}

type ConnectRequest struct {
	ProtocolID uint64
	messageHeader
}

func (r *ConnectRequest) WriteTo(w io.Writer) (int64, error) {
	return 0, binary.Write(w, binary.BigEndian, r)
}

func createConnectRequest() *ConnectRequest {
	req := new(ConnectRequest)
	req.ProtocolID = magicProtocolID
	req.Action = actionConnect
	req.TransactionID = uint32(rand.Int31())
	return req
}

type ConnectResponse struct {
	messageHeader
	ConnectionID uint64
}

func parseConnectResponse(r []byte) *ConnectResponse {
	res := new(ConnectResponse)
	res.Action = binary.BigEndian.Uint32(r[0:4])
	res.TransactionID = binary.BigEndian.Uint32(r[4:8])
	res.ConnectionID = binary.BigEndian.Uint64(r[8:])
	return res
}

type AnnounceRequest struct {
	ConnectionID uint64
	messageHeader
	InfoHash   [20]byte
	PeerID     [20]byte
	Downloaded uint64
	Left       uint64
	Uploaded   uint64
	Event      Event
	IPAddress  uint32
	Key        uint32
	NumWant    int32
	Port       uint16
}

func createAnnounceRequest(connectionID uint64, infoHash [20]byte, left uint64) *AnnounceRequest {
	// default values as per https://www.bittorrent.org/beps/bep_0015.html
	req := new(AnnounceRequest)
	req.ConnectionID = connectionID
	req.Action = actionAnnounce
	req.TransactionID = uint32(rand.Int31())
	req.InfoHash = infoHash
	req.PeerID = GetPeerID()
	req.Downloaded = 0
	req.Left = left
	req.Uploaded = 0
	req.Event = EventNone
	req.IPAddress = 0
	req.Key = rand.Uint32()
	req.NumWant = -1
	req.Port = 6881
	return req
}

func (r *AnnounceRequest) WriteTo(w io.Writer) (int64, error) {
	return 0, binary.Write(w, binary.BigEndian, r)
}

type udpAnnounceResponse struct {
	messageHeader
	Interval uint32
	Leechers uint32
	Seeders  uint32
}

type AnnounceResponse struct {
	messageHeader
	Interval uint32
	Leechers uint32
	Seeders  uint32
	Peers    []*net.TCPAddr
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

func (t *UDPTracker) Connect(torrent torrent.Torrent) *ConnectResponse {
	udpAddr, err := net.ResolveUDPAddr("udp4", t.destination)
	if err != nil {
		fmt.Println("resolve UDP address error:", err)
		return nil
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		t.log.Error().Msgf("%s%v", "create UDP connection error", err)
		return nil
	}
	defer conn.Close()

	req := createConnectRequest()
	_, err = req.WriteTo(conn)

	if err != nil {
		t.log.Error().Msgf("%s%v", "send connect request error", err)
		return nil
	}

	resp := make([]byte, 16)
	_, err = bufio.NewReader(conn).Read(resp)

	if err != nil {
		t.log.Error().Msgf("%s%v", "read connect response error", err)
		return nil
	}

	connResp := parseConnectResponse(resp)
	t.Announce(conn, connResp.ConnectionID, torrent)
	return connResp
}

// TODO: pass connID through context
func (t *UDPTracker) Announce(conn *net.UDPConn, connectionID uint64, torrent torrent.Torrent) (*AnnounceResponse, error) {
	req := createAnnounceRequest(connectionID, torrent.InfoHash(), torrent.TorrentSize())
	_, err := req.WriteTo(conn)

	if err != nil {
		t.log.Error().Msgf("%s%v", "send announce request error", err)
		return nil, err
	}

	data := make([]byte, 1024)
	bytesRead, err := bufio.NewReader(conn).Read(data)

	if err != nil {
		t.log.Error().Msgf("%s%v", "read announce response error", err)
		return nil, err
	}

	res, peers, err := t.parseAnnounceResponse(data, bytesRead)

	if err != nil {
		t.log.Error().Msgf("%s%v", "decode announce request error", err)
		return nil, err
	}

	fmt.Println(res, peers)

	return nil, nil
}

func (t *UDPTracker) parseAnnounceResponse(data []byte, bytesRead int) (*udpAnnounceResponse, []*net.TCPAddr, error) {
	var res udpAnnounceResponse
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &res)

	if err != nil {
		return nil, nil, err
	}

	if res.Action != actionAnnounce {
		return nil, nil, errors.New("invalid action received")
	}

	peers, err := DecodePeerAddrs(data[binary.Size(res):bytesRead])

	if err != nil {
		return nil, nil, err
	}

	return &res, peers, nil
}
