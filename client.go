package main

import (
	"bittorrent-client/torrent"
	"bittorrent-client/tracker"
	"log"
	"os"
)

func main() {
	file, err := os.Open("example.torrent")

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}
	t := torrent.Torrent{}
	t.Unmarshal(file)
	tracker := tracker.New(t.Announce)
	tracker.Connect(t)
}
