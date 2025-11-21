package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
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

	defer func() {
			conn.Close()
			fmt.Println("closed connection")
	}()

	reader := bufio.NewReader(conn)

	baseDir = "."
	args := os.Args

	for i:=0; i<len(args); i++ {
		if args[i] == "--directory" && i+1 < len(args) {
			baseDir = args[i+1]
		}
	}

	for  {
		requestLine, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// client closed connection
				return
			}
			fmt.Println("Error reading request line:", err)
			return
		}
		if requestLine == "\r\n" || requestLine == "\n" {
			continue
		}

		requestArray := strings.Split(requestLine, " ")
		if len(requestArray) < 2 {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nConnection: close\r\nContent-length: 0\r\n\r\n"))
			return
		}
		reqType := requestArray[0]
		fullUrl := requestArray[1]

		fmt.Println("Request line:", requestLine)

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
			parts := strings.SplitN(line, ":",2)
			if len(parts) != 2 {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
		fmt.Println("Headers:\n", headers)

		reqArr := []string{"", fullUrl}
		if reqType == "POST" {
			handlePost(conn, reqArr, headers, reader)
		} else {
			handleGet(conn, reqArr, headers)
		}

		connHdr := ""

		if v,ok := headers["connection"]; ok {
			connHdr = strings.ToLower(v)
		}

		if connHdr == "close" {
			// client asked to close
			fmt.Println("Closed connection")
			conn.Write([]byte("HTTP/1.1 200 OK\r\nConnection: close\r\n\r\n"))
			return
		}
	}
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

	connHdr := includeConnectionClose(headers)
	response := fmt.Sprintf("HTTP/1.1 201 Created\r\n"+
		"%s"+"Content-Length: 0\r\n\r\n",connHdr)
	conn.Write([]byte(response))

}

func handleGet(conn net.Conn,requestArray []string, headers map[string]string) {
	fullUrl := requestArray[1]
	fmt.Println("GET url:", fullUrl)
	fmt.Println("GET headers:", headers)
	if fullUrl == "/" {
		connHdr := includeConnectionClose(headers)
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
		"%s\r\n", connHdr)
		conn.Write([]byte(response))	
		return
	}

	urlParts := strings.Split(fullUrl, "/")
	fmt.Printf("Url Parts: %q\n", urlParts)

	if urlParts[1] == "echo" {
		body:= ""
		if len(urlParts) != 2 {
			body = urlParts[len(urlParts)-1]
		}
		fmt.Println("str:", body)
		if encodingType, ok := headers["accept-encoding"]; ok {
			fmt.Println("Encoding-string:", encodingType)
			encodingSlice := strings.Split(encodingType, ",")
			fmt.Println("Encodings:", encodingSlice)
			gzbody, err := gzipString(body)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				return
			}
			fmt.Println("Gzbody:", gzbody)

			for _, encoding := range encodingSlice {
				if strings.TrimSpace(encoding) == "gzip" {	
					responseHeader := "HTTP/1.1 200 OK\r\n" +
											"Content-Type: text/plain\r\n" +
											"Content-Encoding: gzip\r\n" +
											fmt.Sprintf("Content-Length: %d\r\n\r\n", len(gzbody))
					if _, err := conn.Write([]byte(responseHeader)); err != nil {
						return
					}
					if _, err := conn.Write(gzbody); err != nil {
						// optionally log
					}
					return
				}
			}
			
		}
		contentLength := len(body)
		fmt.Println("Body", contentLength)
		fmt.Println("Content length", contentLength)
		connHdr := includeConnectionClose(headers)
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/plain\r\n"+
			"%s"+
			"Content-Length: %d\r\n\r\n%s",
			connHdr,contentLength, body)
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
		connHdr := includeConnectionClose(headers)
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Content-Type: application/octet-stream\r\n"+
			"%s"+
			"Content-Length: %d\r\n\r\n%s",
			connHdr,fileSize, fileContent)
		conn.Write([]byte(response))
		return
	}

	if urlParts[1] == "user-agent" {
		userAgent := headers["user-agent"]
		body := userAgent

		connHdr := includeConnectionClose(headers)
		resp := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/plain\r\n"+
			"%s"+
			"Content-Length: %d\r\n"+
			"\r\n%s",
			connHdr, len(body), body)

		_, _ = conn.Write([]byte(resp))
		return
	}
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func gzipString(s string)([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err :=  gz.Write([]byte(s))
	if err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func includeConnectionClose(headers map[string]string) string {
	if v,ok := headers["connection"]; ok && strings.ToLower(strings.TrimSpace(v)) == "close" {
		return "Connection: close\r\n"
	}
	return ""
}