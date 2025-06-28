package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		strChan := getLinesChannel(conn)
		for message := range strChan {
			fmt.Printf("%s", message)
		}
		fmt.Printf("\n")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	b := make([]byte, 8)
	currentLine := ""
	strChan := make(chan string)
	go func() {
		defer f.Close()
		defer close(strChan)

		for {
			n, err := f.Read(b)
			if err != nil {
				if err != io.EOF {
					log.Fatalf("error reading message: %v", err)
				}
				if currentLine != "" {
					strChan <- currentLine
				}
				return
			}

			if n == 0 {
				continue
			}

			chunk := string(b[:n])
			sep := strings.Split(chunk, "\n")
			currentLine += sep[0]
			if len(sep) > 1 {
				strChan <- currentLine
				currentLine = sep[1]
			}

		}
	}()
	return strChan
}
