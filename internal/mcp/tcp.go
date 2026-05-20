package mcp

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

// TCPTransport implements Transport over a net.Conn for MCP JSON-RPC.
// Messages are newline-delimited JSON (NDJSON), same as stdio.
type TCPTransport struct {
	conn   net.Conn
	reader *bufio.Reader
	mu     sync.Mutex
}

// NewTCPTransport wraps an existing net.Conn for MCP use.
func NewTCPTransport(conn net.Conn) *TCPTransport {
	return &TCPTransport{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

// Read returns the next line (newline-delimited) from the TCP connection.
func (t *TCPTransport) Read() ([]byte, error) {
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return line, nil
}

// Write sends data to the TCP connection, appending a newline if absent.
func (t *TCPTransport) Write(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	_, err := t.conn.Write(data)
	return err
}

// Close closes the underlying TCP connection.
func (t *TCPTransport) Close() error {
	return t.conn.Close()
}

// ServeTCP listens on the given address and spawns a new MCP Server for each
// accepted connection. The serverTemplate MUST be pre-initialized (mesh loaded
// once); ServeTCP only swaps the transport per connection.
func ServeTCP(addr string, serverTemplate *Server) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("tcp listen %s: %w", addr, err)
	}
	defer ln.Close()

	fmt.Fprintf(os.Stderr, "mimic: listening on %s\n", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "mimic: tcp accept error: %v\n", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			transport := NewTCPTransport(c)
			server := serverTemplate.WithTransport(transport)
			if err := server.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "mimic: tcp session error: %v\n", err)
			}
		}(conn)
	}
}
