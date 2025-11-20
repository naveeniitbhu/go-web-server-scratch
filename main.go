package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var baseDir string

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		panic(err)
	}
	defer l.Close()
	
	fmt.Println("Listening on port: 4221")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error acception connection", err.Error())
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	request, err := reader.ReadString('\n')
	// ReadString('\n') reads until the first \n, so you get just the request 
	// line: "GET /echo/abc HTTP/1.1\r\n"

	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	fmt.Printf("Received request: %s", request)
	requestArray := strings.Split(request, " ")
	fullUrl := requestArray[1]
	reqType := requestArray[0]
	fmt.Printf("Url Path: %q\n", fullUrl)
	fmt.Printf("Request type: %q\n", reqType)

	baseDir = "."
	args := os.Args

	for i:=0; i<len(args); i++ {
		if args[i] == "--directory" && i+1 < len(args) {
			baseDir = args[i+1]
		}
	}

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("error reading header:", err)
			return
		}
		// End of headers on blank line
		if line == "\r\n" || line == "\n"{
			break
		}
		line = strings.TrimRight(line, "\r\n")
		// Spit header into key:value pair
		parts := strings.SplitN(line, ":",2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		headers[key] = value
	}
	fmt.Println("Remaining headers:\n", headers)

	if reqType == "POST" {
		handlePost(conn, requestArray, headers, reader)
		return
	}

	if fullUrl == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))	
		return
	}

	urlParts := strings.Split(fullUrl, "/")
	fmt.Printf("Url Parts: %q\n", urlParts)

	if urlParts[1] == "echo" {
		body:= ""
		if len(urlParts) != 2 {
			body = urlParts[len(urlParts)-1]
		}
		contentLength := len(body)
		fmt.Println(body, contentLength)
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, body)
		conn.Write([]byte(response))
		return
	}

	if urlParts[1] == "files" {
		if len(urlParts) < 3 {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			return
		}
		fmt.Println("Base-dir", baseDir)
		filePath := baseDir + urlParts[2]
		isFilePresent := fileExists(filePath)
		if !isFilePresent {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}
		file, err := os.Open(filePath)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}

		fileInfo, _ := file.Stat()
		fileContent, err := io.ReadAll(file)

		if err != nil {
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}

		fmt.Println("File-Content", fileContent)

		fileSize := fileInfo.Size()
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", fileSize, fileContent)
		conn.Write([]byte(response))
		return
	}

	if urlParts[1] == "user-agent" {
		userAgent := headers["user-agent"]
		body := userAgent

		resp := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n%s", len(body), body)

		_, _ = conn.Write([]byte(resp))
		return
	}
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func handlePost(conn net.Conn,requestArray []string, headers map[string]string, reader *bufio.Reader) {
	fullUrl := requestArray[1]
	fmt.Println("POST url:", fullUrl)
	fmt.Println("POST headers:", headers)

	clrStr, ok := headers["content-length"]
	if !ok {
		conn.Write([]byte("HTTP/1.1 411  Length Required\r\nContent-length: 0\r\n\r\n"))
		return
	}
	fmt.Println("clrStr:", clrStr)
	n, err := strconv.ParseInt(clrStr,10,64)

	if err != nil || n < 0 {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-length: 0\r\n\r\n"))
		return
	}
	urlParts := strings.Split(fullUrl, "/")
	fileName := urlParts[2]
	filePath :=baseDir+fileName
	fmt.Println("File path:", filePath)

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("create error:", err)
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\nContent-Length: 0\r\n\r\n"))
		return
	}

	defer file.Close()
	defer func() {
		if err := file.Close(); err != nil {
				fmt.Println("file close error:", err)
		}
	}()
	written, err := io.CopyN(file, reader, n)
	if err != nil {
    // io.EOF before n bytes or other error
    if errors.Is(err, io.EOF) || written != n {
        conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n"))
    } else {
        conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\nContent-Length: 0\r\n\r\n"))
    }
    return
	}

	// Success — respond (201 Created or 200 OK)
	conn.Write([]byte("HTTP/1.1 201 Created\r\nContent-Length: 0\r\n\r\n"))

}