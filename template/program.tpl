{{- /*gotype: gopawn/internal/msg.GherkinDocument_Program*/ -}}
package main

import (
    "flag"
    "fmt"
    "os"
    "strings"
)

func main() {
    flag.Usage = usage
    ops := parseOptionsFromFlag()
    if ops.showHelp {
        usage()
        os.Exit(2)
    }

    fmt.Println("Hello World")
}

func parseOptionsFromFlag() options {
    ops := options{}
    flag.BoolVar(&ops.showHelp, "help", ops.showHelp, "show usage ")
    flag.Parse()
    return ops
}

func usage() {
    _, _ = fmt.Fprintln(os.Stderr, "Usage of {{.Name}}\n")
    _, _ = fmt.Fprintln(os.Stderr, strings.TrimSpace(`{{.Description}}`))
    _, _ = fmt.Fprintln(os.Stderr, "\n\n"))
    flag.PrintDefaults()
}

type options struct {
    showHelp bool
}
