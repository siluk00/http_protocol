package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	rUDPAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("Couldn't resolve UDP address: %v\n", err)
	}

	udpConn, err := net.DialUDP("udp", nil, rUDPAddr)
	if err != nil {
		log.Fatalf("Couldn't establish UDP connectio: %v\n", err)
	}
	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Couldn't read message: %v\n", err)
		}

		_, err = udpConn.Write([]byte(message))
		if err != nil {
			log.Printf("Couldn't write message: %v\n", err)
		}
	}
}
