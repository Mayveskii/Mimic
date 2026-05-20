// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Mayveskii/Mimic/internal/projectmap"
)

func main() {
	pm, err := projectmap.OpenOrCreate("/home/cisco/mimic")
	if err != nil {
		log.Fatal(err)
	}
	defer pm.Close()

	fmt.Println("Indexing workspace...")
	if err := pm.IndexWorkspace(); err != nil {
		log.Fatal(err)
	}

	stats, _ := pm.Stats()
	b, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(b))

	syms, _ := pm.QuerySymbol("Mesh")
	fmt.Printf("\nSymbols matching 'Mesh': %d\n", len(syms))
	for i, s := range syms {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s (%s) at %s:%d\n", s.Name, s.Type, s.File, s.Line)
	}

	res, _ := pm.SearchText("orchestrator", 5)
	fmt.Printf("\nFTS search 'orchestrator': %d results\n", len(res))
	for i, r := range res {
		if i >= 3 {
			break
		}
		fmt.Printf("  %s (%s)\n", r.Path, r.Lang)
	}
}
