package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	run()
}

func run() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("Error listen to port 42069: %v\n", err)
	}
	defer listener.Close()
	fmt.Println("Listening to port 42069")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v\n", err)
		}
		log.Println("Connection accepted")
		words := getLines(conn)
		fmt.Println(words)
		fmt.Printf("\n")
	}

}

func getLines(f io.ReadCloser) string {
	defer f.Close()
	buffer := make([]byte, 8)
	lines := make([]byte, 0)
	var err error

	for n := 1; err == nil && n > 0; n, err = f.Read(buffer) {
		lines = append(lines, buffer[:n]...)
	}
	if err == io.EOF {
		return string(lines)
	}
	if err != nil {
		log.Printf("couldn't read connection: %v\n", err)
		return ""
	}

	return string(lines)
}
