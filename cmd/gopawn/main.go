package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) < 2 {
		log.Errorf("Provide a *.prog file as first argument")
		return
	}

	input := os.Args[1]
	log.Info("working on: ", input)
}
