package graph

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

const engine = "sfdp"

func MakeGraph() {
	if makeFile() {
		runGV()
	}
}

func makeFile() bool {
	fmt.Println("Writing graph file...")
	file, err := os.Create("graph.dot")
	if err != nil {
		fmt.Println("Error creating graph:", err)
		return false
	}
	defer file.Close()

	file.WriteString("digraph {\n")
	file.WriteString(`graph [bgcolor="#eeeeee", overlap=prism, dpi=120, size=30, `)
	file.WriteString("ratio=0.6, outputorder=edgesfirst]\n")
	file.WriteString("node [fillcolor=\"#eeeeeecc\", color=\"#ffffff\", style=filled, shape=box]\n")
	for _, r := range Links {
		file.WriteString("\"" + r.From + "\" -> \"" + r.To + "\";\n")
	}
	file.WriteString("}\n")
	return true
}

func runGV() {
	fmt.Println("Running", engine)
	start := time.Now()
	cmd := exec.Command(engine, "-K"+engine, "-o", "graph.svg", "-Tsvg", "graph.dot")
	st, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error while running sfdp: ", err)
	}
	fmt.Println(string(st))
	fmt.Println("time: ", time.Since(start))
	fmt.Println(engine, "finished.")
}
