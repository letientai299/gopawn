package composer

import (
	"bytes"
	"testing"
)

func TestWriteProgram(t *testing.T) {
	tests := []struct {
		src     string
		wantOut string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			err := WriteProgram(out, tt.args.prog)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteProgram() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("WriteProgram() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
