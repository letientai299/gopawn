package main

import (
	"gopawn/internal/composer"
	"gopawn/internal/gherkin"
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

	file, err := os.Open(input)
	if err != nil {
		log.Errorf("fail to open file %v, err=%v", input, err)
		return
	}

	doc, err := gherkin.ParseGherkinDocument(file)
	if err != nil {
		log.Errorf("fail to parser document, err=%v", err)
		return
	}

	prog := doc.GetProgram()
	if prog == nil {
		log.Errorf("document contains no program")
		return
	}

	err = composer.WriteProgramTo(prog.Name+".go", prog)
	if err != nil {
		log.Errorf("fail to write program into file, err=%v", err)
		return
	}
}
