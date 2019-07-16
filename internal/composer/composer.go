package composer

import (
	"bytes"
	"gopawn/internal/msg"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

func WriteProgram(out io.Writer, prog *msg.GherkinDocument_Program) error {
	tplSrc, err := getTemplate(tplProgram)
	if err != nil {
		return errors.Wrap(err, "fail to get template src")
	}

	tpl, err := template.New(tplProgram).Parse(tplSrc)
	if err != nil {
		return errors.Wrap(err, "fail to read template "+tplProgram)
	}

	return tpl.Execute(out, prog)
}

func WriteProgramTo(path string, prog *msg.GherkinDocument_Program) error {
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "fail to create new file at "+path)
	}

	if err := WriteProgram(file, prog); err != nil {
		return errors.Wrap(err, "fail to write program")
	}

	return goImports(path)
}

func goImports(path string) error {
	cmd := exec.Command("goimports", "-d", "-e", path)
	var bufOut bytes.Buffer
	var bufErr bytes.Buffer
	cmd.Stderr = &bufErr
	cmd.Stdout = &bufOut

	if err := cmd.Run(); err != nil {
		var errMsg strings.Builder
		errMsg.WriteString("fail to run goimports on " + path)
		errMsg.WriteString(",\nstdout: ")
		errMsg.Write(bufOut.Bytes())
		errMsg.WriteString(",\nstderr:")
		errMsg.Write(bufErr.Bytes())

		return errors.Wrap(err, errMsg.String())
	}

	return nil
}
