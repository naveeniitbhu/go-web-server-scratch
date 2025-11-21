# 🚀 Minimal HTTP Server in Go

A lightweight HTTP/1.1 server built from scratch.  
Supports persistent connections, gzip compression, dynamic headers, static files, uploads, and parallel request handling.

## 📦 Features

- 🔁 Persistent HTTP/1.1 connections (keep-alive)
- 🗃 Static file reads (`GET /files/<name>`)
- 📤 File upload via POST (`POST /files/<name>`)
- 🗣 Echo endpoint (`/echo/<text>`)
- 📱 User-Agent reflection (`/user-agent`)
- 🌀 Gzip compression if requested (`Accept-Encoding: gzip`)
- ⚠️ Proper error handling (400/404/411/500)
- ⚡ Handles multiple connections concurrently

## 🛠 Installation

1. Install **Go 1.24+**

   ```sh
   go version
   ```

2. Run the program:

   ```sh
   ./your_program.sh
   ```

   This script builds and starts the server implemented in `app/main.go`.

## 📡 Understanding `curl -v`

Example request:

```sh
curl -v http://localhost:4221/user-agent   -H "User-Agent: blueberry/raspberry-blueberry"
```

Curl sends:

```
GET /user-agent HTTP/1.1
Host: localhost:4221
User-Agent: blueberry/raspberry-blueberry
Accept: */*
```

Your server must return the `User-Agent` header value in the response body.

## 🔥 Stress Testing via Parallel Requests

Use background jobs to simulate load:

```sh
(sleep 3 && curl -v http://localhost:4221/) &
(sleep 3 && curl -v http://localhost:4221/) &
(sleep 3 && curl -v http://localhost:4221/) &
```

Helps test concurrency and keep-alive behavior.

## 🧪 Persistent Connection Testing

### Request 1 (connection stays open)

```sh
curl --http1.1 -v http://localhost:4221/echo/orange
```

### Request 2 (connection closes)

```sh
curl --http1.1 -v http://localhost:4221/   -H "Connection: close"
```

Expected behavior:

- Server keeps connection open after first request
- Server returns **Connection: close** in response to the second request
- Server closes TCP connection afterward

## 📎 HTTP Response Structure

Each HTTP response must contain a blank line (`\r\n\r\n`) between headers and body:

```
HTTP/1.1 200 OK\r\n
Content-Type: text/plain\r\n
Content-Length: 6\r\n
\r\n
orange
```

Without the blank line, clients cannot parse your response.

## 📁 Example Endpoints

### Echo

```
GET /echo/banana
Response: banana
```

### User-Agent

```
GET /user-agent
Response: <the User-Agent header>
```

### File Read

```
GET /files/example.txt
```

### File Upload

```
POST /files/new.txt
Content-Length: <size>
<body>
```
