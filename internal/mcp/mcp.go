package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Mayveskii/Mimic/internal/cgo"
	"github.com/Mayveskii/Mimic/internal/orchestrator"
)

// Transport handles JSON-RPC message I/O
type Transport interface {
	Read() ([]byte, error)
	Write([]byte) error
	Close() error
}

// StdioTransport implements Transport over stdin/stdout
type StdioTransport struct {
	reader *bufio.Reader
	mu     sync.Mutex
}

func NewStdioTransport() *StdioTransport {
	return &StdioTransport{reader: bufio.NewReader(os.Stdin)}
}

func (t *StdioTransport) Read() ([]byte, error) {
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return line, nil
}

func (t *StdioTransport) Write(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err := os.Stdout.Write(data)
	return err
}

func (t *StdioTransport) Close() error {
	return nil
}

// JSONRPCRequest is a standard MCP request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse is a standard MCP response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError is an error response
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Server is an MCP server that exposes c-core operations as tools via the orchestrator
type Server struct {
	transport    Transport
	tools        []Tool
	orchestrator *orchestrator.Orchestrator
}

// Tool describes an available operation
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// NewServer creates an MCP server with a default budget
func NewServer(t Transport) *Server {
	s := &Server{
		transport:    t,
		orchestrator: orchestrator.NewOrchestrator(10000.0, 1000000.0),
	}
	s.buildTools()
	return s
}

func (s *Server) buildTools() {
	for _, schema := range DefaultSchemas {
		t := Tool{
			Name:        schema.Name,
			Description: schema.Description,
			InputSchema: schema.InputSchema,
		}
		s.tools = append(s.tools, t)
	}
}

// Run starts the server loop
func (s *Server) Run() error {
	for {
		data, err := s.transport.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(data, &req); err != nil {
			// Non-fatal: skip malformed lines
			continue
		}

		resp := s.handleRequest(req)
		if resp != nil {
			out, err := json.Marshal(resp)
			if err != nil {
				continue
			}
			out = append(out, '\n')
			s.transport.Write(out)
		}
	}
}

func (s *Server) handleRequest(req JSONRPCRequest) *JSONRPCResponse {
	resp := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{"listChanged": true},
			},
			"serverInfo": map[string]interface{}{
				"name":    "mimic",
				"version": "0.1.0",
			},
		}
		return resp

	case "notifications/initialized":
		return nil // No response for notifications

	case "tools/list":
		resp.Result = map[string]interface{}{"tools": s.tools}
		return resp

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &JSONRPCError{Code: -32602, Message: "Invalid params: " + err.Error()}
			return resp
		}

		_, err := cgo.PacketFromToolCall(params.Name, params.Arguments)
		if err != nil {
			resp.Error = &JSONRPCError{Code: -32602, Message: "Unknown tool: " + params.Name}
			return resp
		}

		wr, err := s.orchestrator.Run(params.Name, params.Arguments)
		if err != nil {
			resp.Result = map[string]interface{}{
				"content": []map[string]string{
					{"type": "text", "text": fmt.Sprintf("Orchestrator failed: %s", err.Error())},
				},
				"isError": true,
			}
			return resp
		}

		jsonOut, _ := json.MarshalIndent(wr.FinalOutput, "", "  ")
		resp.Result = map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": string(jsonOut)},
			},
		}
		return resp

	case "ping":
		resp.Result = map[string]interface{}{}
		return resp

	default:
		resp.Error = &JSONRPCError{Code: -32601, Message: "Method not found: " + req.Method}
		return resp
	}
}
