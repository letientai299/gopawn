package gherkin

import (
	"gopawn/internal/msg"
	"reflect"
	"strings"
	"testing"
)

func TestParseFeature(t *testing.T) {
	tests := []struct {
		src         string
		wantFeature *msg.GherkinDocument
		wantErr     bool
	}{
		{
			src: `
# simple
Feature: fizzbuzz
`,
		},
	}
	for _, tc := range tests {
		tt := tc
		name := strings.Split(tt.src, "\n")[0]
		name = strings.TrimPrefix(name, "# ")
		t.Run(name, func(t *testing.T) {
			in := strings.NewReader(tt.src)
			gotFeature, err := ParseGherkinDocument(in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFeature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFeature, tt.wantFeature) {
				t.Errorf("ParseFeature() gotFeature = %v, want %v", gotFeature, tt.wantFeature)
			}
		})
	}
}
