package main

import (
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

type Torrent struct {
	Announce     string
	CreatedBy    string `bencode:"created by"`
	CreationDate uint32 `bencode:"creation date"`
	Encoding     string
	Info         Info
}

func main() {
	file, err := os.Open("puppy.torrent")

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	obj := Torrent{}
	err = bencode.Unmarshal(file, &obj)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(obj.Announce, obj.CreationDate)
}
