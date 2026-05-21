package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mayveskii/Mimic/internal/embed"
	"github.com/Mayveskii/Mimic/internal/mesh"
	"github.com/Mayveskii/Mimic/internal/qdrant"
)

// MeshHandler routes MESH_QUERY/MESH_STATUS/EXECUTE_PATTERN to the mesh registry.
type MeshHandler struct {
	registry     *mesh.MeshRegistry
	embedClient  *embed.Client
	qdrantClient *qdrant.Client
	meshDir      string
}

// NewMeshHandler loads mesh graphs from disk.
func NewMeshHandler(meshDir, embedEndpoint string) *MeshHandler {
	if meshDir == "" {
		return &MeshHandler{}
	}
	reg := mesh.NewRegistry()
	if err := reg.LoadAllGraphs(meshDir); err != nil {
		fmt.Fprintf(os.Stderr, "[mesh] warning: failed to load graphs: %v\n", err)
	}
	registryPath := filepath.Join(meshDir, "..", "registry", "invariant_registry.json")
	if _, err := os.Stat(registryPath); err == nil {
		if err := reg.LoadRegistry(registryPath); err != nil {
			fmt.Fprintf(os.Stderr, "[mesh] warning: failed to load registry: %v\n", err)
		}
	}
	return &MeshHandler{
		registry:     reg,
		embedClient:  embed.NewClient(embedEndpoint),
		qdrantClient: qdrant.NewClient("http://localhost:6333", "binary_mesh_chunks"),
		meshDir:      meshDir,
	}
}

// HandleMeshQuery processes the MESH_QUERY tool call.
func (h *MeshHandler) HandleMeshQuery(args map[string]interface{}) map[string]interface{} {
	if h.registry == nil {
		return meshError("mesh not initialized")
	}

	queryText, _ := args["query"].(string)
	if queryText == "" {
		return meshError("'query' is required")
	}

	domain, _ := args["domain"].(string)
	topK := 5
	if tk, ok := args["topK"].(float64); ok {
		topK = int(tk)
	} else if tk, ok := args["topK"].(int); ok {
		topK = tk
	}
	if topK < 1 || topK > 20 {
		topK = 5
	}

	if !h.embedClient.Health() {
		return meshError("embed service unavailable at " + h.embedClient.Endpoint)
	}

	if domain != "" {
		h.registry.SetDomainFilter(domain)
		defer h.registry.ClearDomainFilter()
	}

	result, err := h.registry.Query(queryText, topK, func(text string) [384]int8 {
		emb, err := h.embedClient.EmbedInt8(text)
		if err != nil {
			return [384]int8{}
		}
		return emb
	}, func(text string) []float32 {
		emb, err := h.embedClient.EmbedFloat32(text)
		if err != nil {
			return nil
		}
		return emb
	}, h.qdrantClient)
	if err != nil {
		return meshError("mesh query failed: " + err.Error())
	}

	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": formatMeshResult(result)},
		},
	}
}

// HandleMeshStatus returns mesh runtime info.
func (h *MeshHandler) HandleMeshStatus() map[string]interface{} {
	if h.registry == nil {
		return meshError("mesh not initialized")
	}

	var domains []string
	var totalSlots int
	for _, m := range h.registry.Maps {
		domains = append(domains, fmt.Sprintf("%s/%s=%d", m.Name, m.Domain, len(m.Slots)))
		totalSlots += len(m.Slots)
	}

	embedOK := h.embedClient != nil && h.embedClient.Health()

	status := map[string]interface{}{
		"maps_loaded":    len(h.registry.Maps),
		"total_slots":    totalSlots,
		"domains":        domains,
		"embed_healthy":  embedOK,
		"embed_endpoint": h.embedClient.Endpoint,
	}

	b, _ := json.MarshalIndent(status, "", "  ")
	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": string(b)}},
	}
}

// HandleExecutePattern runs ActionBytes from a mesh slot via C-core.
func (h *MeshHandler) HandleExecutePattern(args map[string]interface{}) map[string]interface{} {
	if h.registry == nil {
		return meshError("mesh not initialized")
	}

	slotID, _ := args["slot_id"].(string)
	if slotID == "" {
		return meshError("'slot_id' is required")
	}

	slot, sm := h.registry.LookupSlotByID(slotID)
	if slot == nil {
		return meshError("slot not found: " + slotID)
	}

	if len(slot.ActionBytes) == 0 {
		return map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": fmt.Sprintf("Slot %s has empty ActionBytes (no automation sequence). Domain: %s, Invariant: %.128s", slotID, sm.Name, slot.Invariant)},
			},
		}
	}

	// Decode ActionBytes using the embryo binary patch decoder
	decoded, err := mesh.FormatActionBytes(slot.ActionBytes)
	if err != nil {
		// Fallback to hex dump for unknown formats
		actionHex := fmt.Sprintf("%x", slot.ActionBytes)
		if len(actionHex) > 800 {
			actionHex = actionHex[:800] + "... [truncated, total " + fmt.Sprintf("%d", len(slot.ActionBytes)) + " bytes]"
		}
		return map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": fmt.Sprintf("Slot: %s\nDomain: %s\nInvariant: %.200s\n\nActionBytes (unknown format, hex dump):\n%s\n\nNote: Binary patch format detected. Agent should analyze and apply manually via FILE_EDIT/SYS_EXEC as appropriate.", slotID, sm.Name, slot.Invariant, actionHex)},
			},
		}
	}

	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": fmt.Sprintf("Slot: %s\nDomain: %s\nInvariant: %.200s\n\nActionBytes decoded (%d bytes):\n%s", slotID, sm.Name, slot.Invariant, len(slot.ActionBytes), decoded)},
		},
	}
}

func meshError(msg string) map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": msg}},
		"isError": true,
	}
}

func formatMeshResult(result *mesh.MeshResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Best match: %s (similarity=%.3f)\n\n", result.BestMap, result.BestSimilarity))
	for i, r := range result.Results {
		inv := r.Slot.Invariant
		if len(inv) > 200 {
			inv = inv[:197] + "..."
		}
		b.WriteString(fmt.Sprintf("%d. [%s] %s  sim=%.3f  usage=%d\n   %s\n\n",
			i+1, r.MapName, r.Slot.SourceRepo, r.Similarity, r.Slot.UsageCount, inv))
	}
	b.WriteString(fmt.Sprintf("Total slots: %d | Maps searched: %d", result.TotalSlots, result.MapsSearched))
	return b.String()
}

// HandleMeshAutoApply queries mesh and auto-applies the best matching pattern.
func (h *MeshHandler) HandleMeshAutoApply(args map[string]interface{}) map[string]interface{} {
	if h.registry == nil {
		return meshError("mesh not initialized")
	}

	queryText, _ := args["query"].(string)
	if queryText == "" {
		return meshError("query is required")
	}

	domain, _ := args["domain"].(string)
	threshold := 0.5
	if t, ok := args["similarity_threshold"].(float64); ok {
		threshold = t
	}

	if !h.embedClient.Health() {
		return meshError("embed service unavailable")
	}

	if domain != "" {
		h.registry.SetDomainFilter(domain)
		defer h.registry.ClearDomainFilter()
	}

	result, err := h.registry.Query(queryText, 3, func(text string) [384]int8 {
		emb, err := h.embedClient.EmbedInt8(text)
		if err != nil {
			return [384]int8{}
		}
		return emb
	}, func(text string) []float32 {
		emb, err := h.embedClient.EmbedFloat32(text)
		if err != nil {
			return nil
		}
		return emb
	}, h.qdrantClient)
	if err != nil {
		return meshError("mesh query failed: " + err.Error())
	}

	if len(result.Results) == 0 {
		return meshError("no matching patterns found")
	}

	best := result.Results[0]
	if best.Similarity < threshold {
		return map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": fmt.Sprintf("Best match similarity %.3f is below threshold %.2f. Pattern not auto-applied.\n\n%s", best.Similarity, threshold, formatMeshResult(result))},
			},
		}
	}

	if len(best.Slot.ActionBytes) == 0 {
		return map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": fmt.Sprintf("Best match has no ActionBytes (no automation sequence). Similarity=%.3f\n\n%s", best.Similarity, formatMeshResult(result))},
			},
		}
	}

	decoded, err := mesh.FormatActionBytes(best.Slot.ActionBytes)
	if err != nil {
		return meshError("decode ActionBytes failed: " + err.Error())
	}

	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": fmt.Sprintf("AUTO-APPLIED pattern (sim=%.3f):\nDomain: %s\nSlot: %s\nInvariant: %.200s\n\nDecoded ActionBytes:\n%s", best.Similarity, best.MapName, best.Slot.ID, best.Slot.Invariant, decoded)},
		},
	}
}
