package main

import (
	"fmt"
	"net"
)

func main() {
	// pattern.Flags |= hyperscan.SomLeftMost

	storage := NewMongoStorage("localhost", 27017, "testing")
	err := storage.Connect(nil)
	if err != nil {
		panic(err)
	}

	importer := NewPcapImporter(storage, net.ParseIP("10.10.10.10"))

	sessionId, err := importer.ImportPcap("capture_00459_20190627165500.pcap")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(sessionId)
	}

}
