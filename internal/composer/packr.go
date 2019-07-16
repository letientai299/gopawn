package composer

import (
	"github.com/gobuffalo/packr/v2"
)

// file under the templates dir
const (
	tplProgram = "program.tpl"
)

//go:generate packr2 clean
//go:generate packr2
var box = packr.New("default", "../../template")

func getTemplate(name string) (string, error) {
	return box.FindString(name)
}
