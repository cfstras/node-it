package main 

import (
	"fmt"
	"github.com/cfstras/node-it/graph"
	"os"
	"encoding/json"
	"bytes"
)

var MagicString = []byte("node-it savefile v01\n")

func main() {
	load()
	
	//initial seeding
	//graph.Failed = append(graph.Failed, "gaming")
	
	fmt.Println("starting")
	graph.Start()
	fmt.Println("finished run, saving data")
	save()
	fmt.Println("exiting happily.")
}

func save() {
	file, err := os.Create("saves.json")
	if err != nil {
		fmt.Println("Error creating savefile:",err)
		return
	}
	defer file.Close()
	file.Write(MagicString)
	enc := json.NewEncoder(file)
	
	err = enc.Encode(graph.Subs)
	if err != nil {
		fmt.Println("Error saving subs:",err)
	}
	err = enc.Encode(graph.Links)
	if err != nil {
		fmt.Println("Error saving links:",err)
	}
	err = enc.Encode(graph.Failed)
	if err != nil {
		fmt.Println("Error saving failed:",err)
	}
	err = enc.Encode(graph.Read)
	if err != nil {
		fmt.Println("Error saving read:",err)
	}
}

func load() {
	file, err := os.Open("saves.json")
	if err != nil {
		fmt.Println("Error opening savefile:",err)
		return
	}
	defer file.Close()
	str := make([]byte, len(MagicString))
	file.Read(str)
	if !bytes.Equal(str,MagicString) {
		fmt.Println("Error: Magic string mismatch. found:",string(str))
		fmt.Println("expected:",string(MagicString))
	}
	enc := json.NewDecoder(file)
	
	err = enc.Decode(&graph.Subs)
	if err != nil {
		fmt.Println("Error loading subs:",err)
	}
	err = enc.Decode(&graph.Links)
	if err != nil {
		fmt.Println("Error loading links:",err)
	}
	err = enc.Decode(&graph.Failed)
	if err != nil {
		fmt.Println("Error loading failed:",err)
	}
	err = enc.Decode(&graph.Read)
	if err != nil {
		fmt.Println("Error loading read:",err)
	}
}
