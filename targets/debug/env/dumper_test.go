package env

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStdoutDumper(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	testCases := map[string]struct {
		writer io.Writer
	}{
		"stdout": {writer: os.Stdout},
		"buffer": {writer: &buf},
	}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dumper := NewStdoutDumper(tt.writer)
			assert.NotNil(t, dumper)
		})
	}
}

func TestStdoutDumper_Dump(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		envs      map[string]string
		expResult string
	}{
		"nil map":   {envs: nil},
		"empty map": {envs: map[string]string{}},
		"envs map": {
			envs: map[string]string{
				"TEST": "test",
				"SOME": "other",
			},
			expResult: "TEST=test\nSOME=other\n",
		},
	}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			dumper := NewStdoutDumper(&buf)
			err := dumper.Dump(tt.envs)
			require.NoError(t, err)
			assert.Equal(t, tt.expResult, buf.String())
		})
	}
}

func TestNewFileDumper(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		filename string
		expErr   string
	}{
		"empty": {filename: "", expErr: "filename must be provided"},
		"ok":    {filename: "test.txt"},
	}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dumper, err := NewFileDumper(tt.filename)
			if tt.expErr != "" {
				assert.Errorf(t, err, tt.expErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dumper)
			}
		})
	}
}

func TestFileDumper_Dump(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		envs      map[string]string
		filename  string
		expResult string
		expErr    string
	}{
		"nil map":   {envs: nil, filename: "/tmp/TestFileDumper_Dump.nil.txt"},
		"empty map": {envs: map[string]string{}, filename: "/tmp/TestFileDumper_Dump.empty.txt"},
		"permission denied": {
			envs: map[string]string{
				"TEST": "test",
				"SOME": "other",
			},
			filename: "/etc/TestFileDumper_Dump.denied.txt",
			expErr:   "failed to create file /etc/TestFileDumper_Dump.denied.txt: open /etc/TestFileDumper_Dump.denied.txt: permission denied",
		},
		"envs map": {
			envs: map[string]string{
				"TEST":    "test",
				"SOME_OF": "other",
			},
			filename:  "/tmp/TestFileDumper_Dump.envs.txt",
			expResult: "TEST=test\nSOME_OF=other\n",
		},
	}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dumper, err := NewFileDumper(tt.filename)
			require.NoError(t, err)
			err = dumper.Dump(tt.envs)
			if tt.expErr != "" {
				assert.EqualError(t, err, tt.expErr)
			} else {
				require.NoError(t, err)
				data, err := ioutil.ReadFile(tt.filename)
				require.NoError(t, err)
				assert.Equal(t, tt.expResult, string(data))
				err = os.Remove(tt.filename)
				require.NoError(t, err)
			}
		})
	}
}
