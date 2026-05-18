package mcp

// ToolSchema определяет JSON Schema для инструмента, совместимый с OpenAI function calling
// и MCP spec. Это решает проблему: модель не знает какие параметры нужны.
type ToolSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// DefaultSchemas содержит схемы для всех реализованных инструментов.
// Генерировано из анализа core/ops.c аргументов.
var DefaultSchemas = []ToolSchema{
	// System Operations
	{
		Name:        "SYS_FILE_EXISTS",
		Description: "Check if a file or directory exists. Returns boolean. Safe and readonly.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute or relative path to check",
				},
			},
			"required": []string{"path"},
		},
	},
	{
		Name:        "SYS_FILE_READ",
		Description: "Read file contents by path. No file descriptor needed. Supports optional limit and offset. Safe and readonly.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File path to read",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Max bytes to read. Default: 4096 (entire file if smaller). 0 = unlimited",
					"default":     4096,
				},
				"offset": map[string]interface{}{
					"type":        "integer",
					"description": "Byte offset to start reading from. Default: 0",
					"default":     0,
				},
			},
			"required": []string{"path"},
		},
	},
	{
		Name:        "SYS_DIR_CREATE",
		Description: "Create a directory. Supports recursive creation. Idempotent (succeeds if already exists).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory path to create",
				},
				"mode": map[string]interface{}{
					"type":        "integer",
					"description": "Permission mode in octal (e.g. 0755). Default: 0755",
					"default":     0755,
				},
				"recursive": map[string]interface{}{
					"type":        "boolean",
					"description": "Create parent directories if they don't exist",
					"default":     false,
				},
			},
			"required": []string{"path"},
		},
	},
	{
		Name:        "SYS_DIR_REMOVE",
		Description: "Remove a directory. DANGEROUS: can delete non-empty directories if recursive=true. Requires explicit permission for recursive mode.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory path to remove",
				},
				"recursive": map[string]interface{}{
					"type":        "boolean",
					"description": "Remove non-empty directories recursively. WARNING: irreversible data loss",
					"default":     false,
				},
			},
			"required": []string{"path"},
		},
	},
	{
		Name:        "SYS_FILE_COPY",
		Description: "Copy a file from source to destination. Overwrites destination if it exists.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"src": map[string]interface{}{
					"type":        "string",
					"description": "Source file path",
				},
				"dst": map[string]interface{}{
					"type":        "string",
					"description": "Destination file path",
				},
			},
			"required": []string{"src", "dst"},
		},
	},
	{
		Name:        "SYS_FILE_MOVE",
		Description: "Move or rename a file. Falls back to copy+delete if rename fails.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"src": map[string]interface{}{
					"type":        "string",
					"description": "Source file path",
				},
				"dst": map[string]interface{}{
					"type":        "string",
					"description": "Destination file path",
				},
			},
			"required": []string{"src", "dst"},
		},
	},
	{
		Name:        "SYS_FILE_DELETE",
		Description: "Delete a file. DANGEROUS: irreversible. Creates backup before deletion if possible.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File path to delete",
				},
			},
			"required": []string{"path"},
		},
	},
	{
		Name:        "SYS_CHMOD",
		Description: "Change file permissions. Blocks setuid/setgid for security.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File or directory path",
				},
				"mode": map[string]interface{}{
					"type":        "integer",
					"description": "Permission mode in octal (e.g. 0644 for rw-r--r--)",
				},
			},
			"required": []string{"path", "mode"},
		},
	},
	{
		Name:        "SYS_ENV_GET",
		Description: "Get an environment variable value. Returns empty string if not set.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Environment variable name",
				},
			},
			"required": []string{"name"},
		},
	},
	{
		Name:        "SYS_ENV_SET",
		Description: "Set an environment variable. DANGEROUS: affects process environment.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Environment variable name",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "Value to set",
				},
			},
			"required": []string{"name", "value"},
		},
	},
	{
		Name:        "SYS_EXEC",
		Description: "Execute a shell command. CRITICAL DANGEROUS: arbitrary code execution. Requires 2-vote verification. Use with extreme caution.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"cmd": map[string]interface{}{
					"type":        "string",
					"description": "Shell command to execute",
				},
			},
			"required": []string{"cmd"},
		},
	},

	// I/O Operations
	{
		Name:        "IO_OPEN",
		Description: "Open a file for reading or writing. Returns file descriptor.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File path to open",
				},
				"mode": map[string]interface{}{
					"type":        "string",
					"description": "Open mode: 'r' (read), 'w' (write), 'a' (append), 'rw' (read-write)",
					"enum":        []string{"r", "w", "a", "rw"},
					"default":     "r",
				},
			},
			"required": []string{"path"},
		},
	},
	{
		Name:        "IO_CLOSE",
		Description: "Close a file descriptor. FD must be valid and previously opened.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"fd": map[string]interface{}{
					"type":        "integer",
					"description": "File descriptor to close",
				},
			},
			"required": []string{"fd"},
		},
	},
	{
		Name:        "IO_READ",
		Description: "Read data from a file descriptor.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"fd": map[string]interface{}{
					"type":        "integer",
					"description": "File descriptor to read from",
				},
				"length": map[string]interface{}{
					"type":        "integer",
					"description": "Number of bytes to read",
				},
			},
			"required": []string{"fd", "length"},
		},
	},
	{
		Name:        "IO_WRITE",
		Description: "Write data to a file descriptor.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"fd": map[string]interface{}{
					"type":        "integer",
					"description": "File descriptor to write to",
				},
				"data": map[string]interface{}{
					"type":        "string",
					"description": "Data to write",
				},
			},
			"required": []string{"fd", "data"},
		},
	},
	{
		Name:        "IO_SEEK",
		Description: "Seek to a position in a file. whence: 0=SET, 1=CUR, 2=END.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"fd": map[string]interface{}{
					"type":        "integer",
					"description": "File descriptor",
				},
				"offset": map[string]interface{}{
					"type":        "integer",
					"description": "Byte offset",
				},
				"whence": map[string]interface{}{
					"type":        "integer",
					"description": "Seek origin: 0=SEEK_SET, 1=SEEK_CUR, 2=SEEK_END",
					"enum":        []int{0, 1, 2},
					"default":     0,
				},
			},
			"required": []string{"fd", "offset"},
		},
	},

	// Build Operations
	{
		Name:        "BUILD_COMPILE",
		Description: "Compile a target using make or equivalent. Captures stdout/stderr.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Make target name (e.g. 'all', 'mimic')",
				},
				"flags": map[string]interface{}{
					"type":        "string",
					"description": "Additional compiler flags",
				},
			},
		},
	},
	{
		Name:        "BUILD_LINK",
		Description: "Link object files into executable or library.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"inputs": map[string]interface{}{
					"type":        "string",
					"description": "Object files or libraries to link (space-separated)",
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Output file name",
				},
			},
			"required": []string{"inputs", "output"},
		},
	},
	{
		Name:        "BUILD_TEST",
		Description: "Run tests (Go test by default). Can filter by pattern and specify working directory.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filter": map[string]interface{}{
					"type":        "string",
					"description": "Test name pattern (e.g. 'TestValidation')",
				},
				"timeout_ms": map[string]interface{}{
					"type":        "integer",
					"description": "Timeout in milliseconds. Default: 30000",
					"default":     30000,
				},
				"dir": map[string]interface{}{
					"type":        "string",
					"description": "Working directory for tests",
				},
			},
		},
	},
	{
		Name:        "BUILD_DEPLOY",
		Description: "Deploy an artifact. CRITICAL DANGEROUS: affects production. Requires 2-vote verification.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Deployment target (server, cluster, environment)",
				},
				"version": map[string]interface{}{
					"type":        "string",
					"description": "Version tag or commit hash to deploy",
				},
			},
			"required": []string{"target"},
		},
	},
	{
		Name:        "BUILD_CLEAN",
		Description: "Clean build artifacts. DANGEROUS: deletes generated files.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Clean target (e.g. 'all', 'test-artifacts')",
				},
			},
		},
	},

	// Git Operations
	{
		Name:        "GIT_STATUS",
		Description: "Get git working tree status. Safe and readonly.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
	{
		Name:        "GIT_DIFF",
		Description: "Show git diff statistics. Safe and readonly.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
	{
		Name:        "GIT_ADD",
		Description: "Stage files for commit. If no path specified, stages all changes (git add -A).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File path to stage (optional, defaults to all)",
				},
			},
		},
	},
	{
		Name:        "GIT_COMMIT",
		Description: "Create a git commit. DANGEROUS: creates history entry.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "Commit message",
				},
			},
		},
	},
	{
		Name:        "GIT_CHECKOUT",
		Description: "Switch to a branch or commit. DANGEROUS: can discard uncommitted changes.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"branch": map[string]interface{}{
					"type":        "string",
					"description": "Branch name or commit hash to checkout",
				},
			},
			"required": []string{"branch"},
		},
	},
	{
		Name:        "GIT_BRANCH",
		Description: "Create or list branches. If name is specified, creates new branch.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "New branch name (optional, lists branches if omitted)",
				},
			},
		},
	},

	// Network Operations
	{
		Name:        "NET_HTTP_GET",
		Description: "Make an HTTP GET request. Returns response body.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "URL to fetch",
				},
			},
			"required": []string{"url"},
		},
	},
	{
		Name:        "NET_HTTP_POST",
		Description: "Make an HTTP POST request. Returns response body.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "URL to post to",
				},
				"data": map[string]interface{}{
					"type":        "string",
					"description": "POST body data",
				},
			},
			"required": []string{"url"},
		},
	},
	{
		Name:        "NET_TCP_CLOSE",
		Description: "Close a TCP socket by file descriptor.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"fd": map[string]interface{}{
					"type":        "integer",
					"description": "Socket file descriptor",
				},
			},
			"required": []string{"fd"},
		},
	},

	// Process Operations
	{
		Name:        "PROC_SPAWN",
		Description: "Spawn a new process. DANGEROUS: resource consumption. Requires explicit permission.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"cmd": map[string]interface{}{
					"type":        "string",
					"description": "Command to execute",
				},
			},
			"required": []string{"cmd"},
		},
	},
	{
		Name:        "PROC_WAIT",
		Description: "Wait for a process to complete. Safe.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pid": map[string]interface{}{
					"type":        "integer",
					"description": "Process ID to wait for",
				},
			},
			"required": []string{"pid"},
		},
	},
	{
		Name:        "PROC_KILL",
		Description: "Kill a process. CRITICAL DANGEROUS: terminates running process.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pid": map[string]interface{}{
					"type":        "integer",
					"description": "Process ID to kill",
				},
				"signal": map[string]interface{}{
					"type":        "integer",
					"description": "Signal number (default: 9 = SIGKILL)",
					"default":     9,
				},
			},
			"required": []string{"pid"},
		},
	},
	{
		Name:        "PROC_SIGNAL",
		Description: "Send a signal to a process. DANGEROUS: can alter process behavior.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pid": map[string]interface{}{
					"type":        "integer",
					"description": "Process ID",
				},
				"signal": map[string]interface{}{
					"type":        "integer",
					"description": "Signal number",
					"default":     15,
				},
			},
			"required": []string{"pid"},
		},
	},

	// Utility Operations
	{
		Name:        "HASH_SHA256",
		Description: "Compute SHA256 hash of data. Safe and readonly.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data": map[string]interface{}{
					"type":        "string",
					"description": "Data to hash",
				},
			},
			"required": []string{"data"},
		},
	},
	{
		Name:        "HASH_MD5",
		Description: "Compute MD5 hash of data. Safe and readonly.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data": map[string]interface{}{
					"type":        "string",
					"description": "Data to hash",
				},
			},
			"required": []string{"data"},
		},
	},
}

// SchemaMap быстрый lookup по имени инструмента
var SchemaMap map[string]ToolSchema

func init() {
	SchemaMap = make(map[string]ToolSchema, len(DefaultSchemas))
	for _, s := range DefaultSchemas {
		SchemaMap[s.Name] = s
	}
}
