package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/Mayveskii/Mimic/internal/cas"
)

// CASTools routes Content-Addressable Store (CAS) pattern tools to the
// internal/cas Store and Assembler.
type CASTools struct {
	store     *cas.Store
	assembler *cas.Assembler
}

// RegisterCASTools wires the CAS pattern tools into the MCP server.
func RegisterCASTools(server *Server, store *cas.Store, assembler *cas.Assembler) {
	server.casHandler = &CASTools{
		store:     store,
		assembler: assembler,
	}
}

// HandleApplyPattern applies a stored pattern to a base tree and returns the
// resulting tree SHA. If tree_sha is omitted the empty tree is used.
func (h *CASTools) HandleApplyPattern(args map[string]interface{}) map[string]interface{} {
	patternSHA, _ := args["pattern_sha"].(string)
	if patternSHA == "" {
		return casError("'pattern_sha' is required")
	}
	targetPath, _ := args["target_path"].(string)
	if targetPath == "" {
		return casError("'target_path' is required")
	}

	params := map[string]string{}
	if raw, ok := args["params"].(map[string]interface{}); ok {
		for k, v := range raw {
			if s, ok := v.(string); ok {
				params[k] = s
			} else {
				params[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	baseTree := treeSHAFromArgs(args)
	newTree, err := h.assembler.Assemble(context.Background(), baseTree, []cas.AssemblyStep{
		cas.ApplyPatternStep{
			PatternSHA: patternSHA,
			TargetPath: targetPath,
			Params:     params,
		},
	})
	if err != nil {
		return casError("apply pattern failed: " + err.Error())
	}

	return casResult(map[string]interface{}{"tree_sha": newTree})
}

// HandleGraftModule copies a source tree into a target path inside a new tree.
func (h *CASTools) HandleGraftModule(args map[string]interface{}) map[string]interface{} {
	sourceTreeSHA, _ := args["source_tree_sha"].(string)
	if sourceTreeSHA == "" {
		return casError("'source_tree_sha' is required")
	}
	targetPath, _ := args["target_path"].(string)
	if targetPath == "" {
		return casError("'target_path' is required")
	}

	steps, err := h.graftSteps(sourceTreeSHA, targetPath)
	if err != nil {
		return casError("graft module failed: " + err.Error())
	}

	baseTree := treeSHAFromArgs(args)
	newTree, err := h.assembler.Assemble(context.Background(), baseTree, steps)
	if err != nil {
		return casError("graft module failed: " + err.Error())
	}

	return casResult(map[string]interface{}{"tree_sha": newTree})
}

func (h *CASTools) graftSteps(treeSHA, prefix string) ([]cas.AssemblyStep, error) {
	entries, err := h.store.ReadTree(treeSHA)
	if err != nil {
		return nil, fmt.Errorf("read tree %s: %w", treeSHA, err)
	}

	var steps []cas.AssemblyStep
	for _, e := range entries {
		switch e.Type {
		case "blob":
			data, err := h.store.ReadBlob(e.SHA)
			if err != nil {
				return nil, fmt.Errorf("read blob %s: %w", e.SHA, err)
			}
			steps = append(steps, cas.WriteFileStep{
				Path:    filepath.Join(prefix, e.Path),
				Content: data,
			})
		case "tree":
			subSteps, err := h.graftSteps(e.SHA, filepath.Join(prefix, e.Path))
			if err != nil {
				return nil, err
			}
			steps = append(steps, subSteps...)
		}
	}
	return steps, nil
}

// HandleEditViaPatch applies a unified diff patch to a file in a base tree. If
// tree_sha is omitted the empty tree is used.
func (h *CASTools) HandleEditViaPatch(args map[string]interface{}) map[string]interface{} {
	filePath, _ := args["file_path"].(string)
	if filePath == "" {
		return casError("'file_path' is required")
	}
	patch, _ := args["patch"].(string)

	baseTree := treeSHAFromArgs(args)
	newTree, err := h.assembler.Assemble(context.Background(), baseTree, []cas.AssemblyStep{
		cas.EditPatchStep{Path: filePath, Patch: patch},
	})
	if err != nil {
		return casError("edit via patch failed: " + err.Error())
	}

	return casResult(map[string]interface{}{"tree_sha": newTree})
}

// HandleWriteFile writes raw content to a path in a base tree and returns the
// new tree SHA. If tree_sha is omitted the empty tree is used.
func (h *CASTools) HandleWriteFile(args map[string]interface{}) map[string]interface{} {
	path, _ := args["path"].(string)
	if path == "" {
		return casError("'path' is required")
	}
	content, _ := args["content"].(string)

	baseTree := treeSHAFromArgs(args)
	newTree, err := h.assembler.Assemble(context.Background(), baseTree, []cas.AssemblyStep{
		cas.WriteFileStep{Path: path, Content: []byte(content)},
	})
	if err != nil {
		return casError("write file failed: " + err.Error())
	}

	return casResult(map[string]interface{}{"tree_sha": newTree})
}

// treeSHAFromArgs returns an explicit tree_sha argument or the empty tree.
func treeSHAFromArgs(args map[string]interface{}) string {
	if sha, _ := args["tree_sha"].(string); sha != "" {
		return sha
	}
	return cas.EmptyTreeSHA
}

// HandleCommitTransition creates a commit object from a tree SHA.
func (h *CASTools) HandleCommitTransition(args map[string]interface{}) map[string]interface{} {
	treeSHA, _ := args["tree_sha"].(string)
	if treeSHA == "" {
		return casError("'tree_sha' is required")
	}
	parentSHA, _ := args["parent_sha"].(string)
	message, _ := args["message"].(string)
	if message == "" {
		return casError("'message' is required")
	}

	commitSHA, err := h.store.StoreCommit(treeSHA, parentSHA, message)
	if err != nil {
		return casError("commit transition failed: " + err.Error())
	}

	return casResult(map[string]interface{}{"commit_sha": commitSHA})
}

// HandleReadBlob returns the contents of a blob object.
func (h *CASTools) HandleReadBlob(args map[string]interface{}) map[string]interface{} {
	sha, _ := args["sha"].(string)
	if sha == "" {
		return casError("'sha' is required")
	}

	data, err := h.store.ReadBlob(sha)
	if err != nil {
		return casError("read blob failed: " + err.Error())
	}

	return casResult(map[string]interface{}{
		"sha":     sha,
		"content": string(data),
	})
}

// HandleReadTree returns the entries of a tree object.
func (h *CASTools) HandleReadTree(args map[string]interface{}) map[string]interface{} {
	sha, _ := args["sha"].(string)
	if sha == "" {
		return casError("'sha' is required")
	}

	entries, err := h.store.ReadTree(sha)
	if err != nil {
		return casError("read tree failed: " + err.Error())
	}

	return casResult(map[string]interface{}{
		"sha":     sha,
		"entries": entries,
	})
}

func casError(msg string) map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": msg}},
		"isError": true,
	}
}

func casResult(payload map[string]interface{}) map[string]interface{} {
	b, _ := json.Marshal(payload)
	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": string(b)}},
	}
}

// isCASTool reports whether name is handled by CASTools.
func isCASTool(name string) bool {
	switch name {
	case "apply_pattern", "graft_module", "edit_via_patch", "write_file",
		"commit_transition", "read_blob", "read_tree":
		return true
	}
	return false
}

// dispatchCASTools routes a CAS tool call to the appropriate handler.
func (s *Server) dispatchCASTools(name string, args map[string]interface{}) map[string]interface{} {
	if s.casHandler == nil {
		return casError("CAS tools not registered")
	}

	switch name {
	case "apply_pattern":
		return s.casHandler.HandleApplyPattern(args)
	case "graft_module":
		return s.casHandler.HandleGraftModule(args)
	case "edit_via_patch":
		return s.casHandler.HandleEditViaPatch(args)
	case "write_file":
		return s.casHandler.HandleWriteFile(args)
	case "commit_transition":
		return s.casHandler.HandleCommitTransition(args)
	case "read_blob":
		return s.casHandler.HandleReadBlob(args)
	case "read_tree":
		return s.casHandler.HandleReadTree(args)
	}

	return casError("unknown CAS tool: " + name)
}
