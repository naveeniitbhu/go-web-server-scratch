## 
GET /index.html HTTP/1.1\r\nHost: localhost:4221\r\nUser-Agent: curl/7.64.1\r\nAccept: _/_\r\n\r\n

// Request line
GET // HTTP method
/index.html // Request target
HTTP/1.1 // HTTP version
\r\n // CRLF that marks the end of the request line

// Headers
Host: localhost:4221\r\n // Header that specifies the server's host and port
User-Agent: curl/7.64.1\r\n // Header that describes the client's user agent
Accept: _/_\r\n // Header that specifies which media types the client can accept
\r\n // CRLF that marks the end of the headers

// Request body (empty)

[]byte("Hello")
Becomes: [72, 101, 108, 108, 111] (ASCII values)

conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")) -- fails
first example worked because you correctly included Content-Length to specify
the size of "Hello Client!\n". Your second example failed because you omitted
this crucial header, leaving the client clueless about where the
response body ends.