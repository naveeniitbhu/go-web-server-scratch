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
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}	
	fmt.Println("Accept connection successfull")

	reader := bufio.NewReader(conn)
	request, err := reader.ReadString('\n')
	// ReadString('\n') reads until the first \n, so you get just the request 
	// line: "GET /echo/abc HTTP/1.1\r\n"

	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	fmt.Printf("Received request: %s", request)

	fullUrl := strings.Split(request, " ")[1]
	fmt.Printf("Url Path: %q\n", fullUrl)

	if fullUrl == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))	
		return
	}

	urlParts := strings.Split(fullUrl, "/")
	fmt.Printf("Url Parts: %q\n", urlParts)

	if len(urlParts) == 2 {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}

	body := urlParts[len(urlParts)-1]
	contentLength := len(body)
	fmt.Println(body, contentLength)

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, body)

	conn.Write([]byte(response))
}


	// response := "HTTP/1.1 200 OK\r\n" +
  //          "Content-Type: text/plain\r\n" +
  //          "Content-Length: 13\r\n" +
  //          "\r\n" +
  //          "Hello, World!"
	// conn.Write([]byte(response))