package main

import (
	_ "embed"
	"github.com/derekmu/lscm"
	"log"
)

//go:embed human.obj
var human string

func main() {
	log.Printf("Parsing mesh")

	mesh := lscm.NewMesh()
	err := mesh.Parse(human)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Running LSCM")

	cm := lscm.NewLSCM(mesh)
	err = cm.Project()
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Writing mesh to ./human.out.obj")

	err = mesh.Write("./human-out.obj")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Done")
}
