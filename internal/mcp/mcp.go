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
	"github.com/Mayveskii/Mimic/internal/tool/exa"
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
	transport         Transport
	tools             []Tool
	orchestrator      *orchestrator.Orchestrator
	workingDir        string
	meshHandler       *MeshHandler
	projectMapHandler *ProjectMapHandler
	exaHandler        *ExaHandler
	planHandler       *PlanHandler
}

// Tool describes an available operation
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
	Group       string                 `json:"group,omitempty"`
}

// NewServer creates an MCP server.
func NewServer(t Transport, workingDir, meshDir, embedEndpoint string) *Server {
	s := &Server{
		transport:         t,
		orchestrator:      orchestrator.NewOrchestrator(10000.0, 1000000.0),
		workingDir:        workingDir,
		meshHandler:       NewMeshHandler(meshDir, embedEndpoint),
		projectMapHandler: NewProjectMapHandler(workingDir, embedEndpoint),
		exaHandler:        NewExaHandler(exa.LoadConfigFromEnv()),
		planHandler:       NewPlanHandler(),
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
			Group:       schema.Group,
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
		var params struct {
			Context string `json:"context,omitempty"`
		}
		_ = json.Unmarshal(req.Params, &params)
		tools := s.tools
		if params.Context != "" {
			schemas := ToolsForContext(params.Context)
			tools = nil
			for _, sc := range schemas {
				tools = append(tools, Tool{
					Name:        sc.Name,
					Description: sc.Description,
					InputSchema: sc.InputSchema,
					Group:       sc.Group,
				})
			}
		}
		resp.Result = map[string]interface{}{"tools": tools}
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

		// Route mesh tools through Go handler (not C-core)
		if params.Name == "MESH_QUERY" {
			resp.Result = s.meshHandler.HandleMeshQuery(params.Arguments)
			return resp
		}
		if params.Name == "MESH_STATUS" {
			resp.Result = s.meshHandler.HandleMeshStatus()
			return resp
		}
		if params.Name == "EXECUTE_PATTERN" {
			resp.Result = s.meshHandler.HandleExecutePattern(params.Arguments)
			return resp
		}

		// Route project map tools
		if params.Name == "PROJECT_MAP_INDEX" {
			resp.Result = s.projectMapHandler.HandleIndex()
			return resp
		}
		if params.Name == "PROJECT_MAP_STATUS" {
			resp.Result = s.projectMapHandler.HandleStatus()
			return resp
		}
		if params.Name == "PROJECT_MAP_QUERY_SYMBOL" {
			resp.Result = s.projectMapHandler.HandleQuerySymbol(params.Arguments)
			return resp
		}
		if params.Name == "PROJECT_MAP_SEARCH_TEXT" {
			resp.Result = s.projectMapHandler.HandleSearchText(params.Arguments)
			return resp
		}
		if params.Name == "WORKSPACE_SYNTHESIZE" {
			resp.Result = s.projectMapHandler.HandleSynthesize()
			return resp
		}
		if params.Name == "MESH_AUTO_APPLY" {
			resp.Result = s.meshHandler.HandleMeshAutoApply(params.Arguments)
			return resp
		}
		if params.Name == "PLAN_GENERATE" {
			resp.Result = s.planHandler.HandleGeneratePlan(params.Arguments)
			return resp
		}

		// Route Exa tools
		if params.Name == "EXA_SEARCH" {
			resp.Result = s.exaHandler.HandleExaSearch(params.Arguments)
			return resp
		}
		if params.Name == "EXA_FETCH" {
			resp.Result = s.exaHandler.HandleExaFetch(params.Arguments)
			return resp
		}
		if params.Name == "MIMIC_RESEARCH" {
			resp.Result = s.exaHandler.HandleMimicResearch(params.Arguments)
			return resp
		}

		// Fast path: direct execution via cgo (bypass orchestrator for simple tool calls)
		pkt, err := cgo.PacketFromToolCall(params.Name, params.Arguments)
		if err != nil {
			resp.Error = &JSONRPCError{Code: -32602, Message: "Unknown tool: " + params.Name}
			return resp
		}

		cr, err := cgo.ExecuteChain([]cgo.Packet{pkt}, 100000.0, 60000.0)
		if err != nil {
			resp.Result = map[string]interface{}{
				"content": []map[string]string{
					{"type": "text", "text": fmt.Sprintf("Execution failed: %s (code=%d)", cr.ErrorMessage, cr.ErrorCode)},
				},
				"isError": true,
			}
			return resp
		}

		resultText := cr.Result
		if resultText == "" {
			jsonOut, _ := json.MarshalIndent(cr, "", "  ")
			resultText = string(jsonOut)
		}
		resp.Result = map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": resultText},
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

// WithTransport clones s with a new transport for TCP multi-client use.
// Handlers (mesh, projectMap, tools, orchestrator) are shared — only the
// transport and per-connection state are replaced.
func (s *Server) WithTransport(t Transport) *Server {
	return &Server{
		transport:         t,
		tools:             s.tools,
		orchestrator:      s.orchestrator,
		workingDir:        s.workingDir,
		meshHandler:       s.meshHandler,
		projectMapHandler: s.projectMapHandler,
		exaHandler:        s.exaHandler,
		planHandler:       s.planHandler,
	}
}
