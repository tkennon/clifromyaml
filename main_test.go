package main

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	//go:embed golden_help.test
	goldenHelp string
)

func TestClifromyaml(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			name:           "no args",
			args:           []string{},
			expectErr:      true,
			expectedOutput: "'clifromyaml': too few arguments; expect 1, but got 0",
		},
		{
			name:           "basic help",
			args:           []string{"-h"},
			expectErr:      false,
			expectedOutput: goldenHelp,
		},
		{
			name:      "dry run",
			args:      []string{"--dry-run", "cli.yaml"},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			cli := NewCLIWithWriter(buf, application{})
			err := cli.clifromyamlCommand.run(tt.args)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}
