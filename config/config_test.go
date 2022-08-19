package config_test

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"hostchecker/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden image")

func TestGet(t *testing.T) {
	testCases := []struct {
		content string
		pass    bool
	}{
		{
			content: `debug: true
port: 9000
targets:
  http:
  - name: example
    url: https://example.com
`,
			pass: true,
		},
		{
			content: `targets:
  http:
  - name: example
    url: https://example.com
    method: POST
    codes: [ 201 ]`,
			pass: true,
		},
		{
			content: `not a yaml file`,
			pass:    false,
		},
		{
			content: ``,
			pass:    false,
		},
		/*
		   		{
		   			content: `sites:
		     - name: example
		       url: https://example.com
		       method: INVALID
		   `,
		   			pass: false,
		   		},
		*/
	}

	for index, tt := range testCases {
		cfg, err := config.Read(bytes.NewBufferString(tt.content))

		if !tt.pass {
			assert.Error(t, err, index)
			continue
		}

		require.NoError(t, err, index)

		var output []byte
		output, err = yaml.Marshal(cfg)
		require.NoError(t, err, index)

		gp := filepath.Join("testdata", fmt.Sprintf("%s_%d_golden.yaml", strings.ToLower(t.Name()), index))
		if *update {
			err = os.WriteFile(gp, output, 0644)
			require.NoError(t, err, index)
		}

		var golden []byte
		golden, err = os.ReadFile(gp)
		require.NoError(t, err, index)

		assert.Equal(t, string(golden), string(output), index)
	}
}
