package tracker

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"sync"
)

type Peer struct {
	PeerID [20]byte
}

var (
	peer       *Peer
	peerIDOnce sync.Once
)

func GetPeerID() [20]byte {
	peerIDOnce.Do(func() {
		buf := make([]byte, 20)
		copy(buf, []byte("-ST001-"))
		rand.Read(buf[7:])
		peer = &Peer{
			PeerID: ([20]byte)(buf),
		}
	})
	return peer.PeerID
}

type PeerAddr struct {
	IP   [4]byte
	Port uint16
}

func (a PeerAddr) TCPAddr() *net.TCPAddr {
	return &net.TCPAddr{IP: a.IP[:], Port: int(a.Port)}
}

func (a *PeerAddr) FromBinary(data []byte) error {
	if len(data) != 6 {
		return errors.New("peer address data must be of length 6")
	}
	return binary.Read(bytes.NewReader(data), binary.BigEndian, a)
}

func DecodePeerAddrs(data []byte) ([]*net.TCPAddr, error) {
	if len(data)%6 != 0 {
		return nil, errors.New("invalid peer list length")
	}
	peersCount := len(data) / 6
	addrs := make([]*net.TCPAddr, 0, peersCount)

	for i := 0; i < len(data); i++ {
		var peerAddr PeerAddr
		err := peerAddr.FromBinary(data[i : i+6])

		if err != nil {
			return nil, err
		}
		addrs = append(addrs, peerAddr.TCPAddr())
	}

	return addrs, nil
}
