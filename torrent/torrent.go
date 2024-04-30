package torrent

import (
	"bytes"
	"crypto/sha1"
	"os"

	"github.com/jackpal/bencode-go"
)

type file struct {
	Length uint64
	Path   []string
}

type info struct {
	PieceLength uint64 `bencode:"piece length"`
	Pieces      string
	Name        string
	Length      uint64
	Files       []file
}

type Torrent struct {
	Announce     string
	CreatedBy    string `bencode:"created by"`
	CreationDate uint32 `bencode:"creation date"`
	Encoding     string
	Info         info
}

func (tf *Torrent) Unmarshal(file *os.File) error {
	return bencode.Unmarshal(file, tf)
}

func (tf *Torrent) InfoHash() [20]byte {
	buf := new(bytes.Buffer)
	bencode.Marshal(buf, tf.Info)
	return sha1.Sum(buf.Bytes())
}

func (tf *Torrent) TorrentSize() uint64 {
	if len(tf.Info.Files) > 0 {
		var sum uint64
		for _, v := range tf.Info.Files {
			sum += v.Length
		}
		return sum
	}
	return tf.Info.Length
}
