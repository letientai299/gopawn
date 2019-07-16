package gherkin

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFeature(t *testing.T) {
	src := `
# should still parse valid gherkin
Program: fizzbuzz program

 A test program.
`
	in := strings.NewReader(src)
	doc, err := ParseGherkinDocument(in)

	assert.NoError(t, err, "don't want err")
	fmt.Println(doc.GetFeature())
	fmt.Println(doc.GetComments())
	fmt.Println(doc.GetProgram())
	fmt.Println(doc.GetProgram().GetDescription())
}
