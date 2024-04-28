package main

import (
	"bittorrent-client/tracker"
	"fmt"
	"log"
	"os"

	bencode "github.com/jackpal/bencode-go"
)

type Info struct {
	PieceLength uint32 `bencode:"piece length"`
	Pieces      string
	Name        string
	Length      uint32
}

type TorrentFile struct {
	Announce     string
	CreatedBy    string `bencode:"created by"`
	CreationDate uint32 `bencode:"creation date"`
	Encoding     string
	Info         Info
}

func main() {
	file, err := os.Open("example.torrent")

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	obj := TorrentFile{}
	err = bencode.Unmarshal(file, &obj)
	if err != nil {
		log.Fatal(err)
	}
	tracker := tracker.New(obj.Announce)
	connectResponse := tracker.Connect()
	fmt.Println(connectResponse)
}
