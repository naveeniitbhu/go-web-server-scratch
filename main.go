package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Println("Listening on port: 4221")
	
	// This is a blocking call. It waits indefinitely until a client actually
	// connects to this server - Test it with curl localhost:4221
	conn, err := l.Accept()

	if err != nil {
		fmt.Println("Error acepting connection: ", err.Error())
		os.Exit(1)
	}	
	fmt.Println("Accept connection successfull")
	// response := "HTTP/1.1 200 OK\r\n" +
  //          "Content-Type: text/plain\r\n" +
  //          "Content-Length: 13\r\n" +
  //          "\r\n" +
  //          "Hello, World!"
	// conn.Write([]byte(response))
	reader := bufio.NewReader(conn)
	request, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}
	fmt.Printf("Received request: %s", request)
	url := strings.Split(request, " ")[1]
	fmt.Printf("Url Path: %s\n", url)
	if url != "/" {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	}
}
