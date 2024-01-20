package main

import (
	"bufio"
	_ "embed"
	"github.com/derekmu/lscm"
	"log"
	"os"
)

//go:embed human.obj
var human string

func main() {
	log.Printf("Parsing mesh")

	mesh := lscm.NewMesh()
	if err := mesh.ParseObj(human); err != nil {
		log.Panic(err)
	}

	log.Printf("Running LSCM")

	cm := lscm.NewLSCM(mesh)
	if err := cm.Project(); err != nil {
		log.Panic(err)
	}

	log.Printf("Writing mesh to ./human-out.obj")

	file, err := os.Create("./human-out.obj")
	if err != nil {
		log.Panic(err)
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Panic(err)
		}
	}(file)
	writer := bufio.NewWriter(file)
	if err = mesh.WriteObj(writer); err != nil {
		log.Panic(err)
	}
	if err = writer.Flush(); err != nil {
		log.Panic(err)
	}

	log.Printf("Done")
}
