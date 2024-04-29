package tracker

import (
	"crypto/rand"
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
