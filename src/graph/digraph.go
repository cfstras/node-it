package graph

import (
	"os"
	"fmt"
	"os/exec"
)

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
	file.WriteString(`graph [splines=spline, overlap=false, fontname="Calibri", dpi=120, size=20, ratio=1.6]`)
	file.WriteString("\n")
	for _, r := range Links {
		file.WriteString(r.From+" -> "+r.To+"\n")
	}
	file.WriteString("}\n")
	return true
}

func runGV() {
	fmt.Println("Running sfdp")
	cmd := exec.Command("sfdp","-o","graph.png","-Tpng","graph.dot")
	st, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error while running sfdp: ", err)
	}
	fmt.Println(string(st))
	fmt.Println("sfdp finished.")
}