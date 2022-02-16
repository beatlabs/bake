package session

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DumpToFile(t *testing.T) {
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
				"TEST_PORT": "9090",
				"SOME_OF":   "other",
				"TEST_HOST": "test",
			},
			filename:  "/tmp/TestFileDumper_Dump.envs.txt",
			expResult: "SOME_OF=other\nTEST_HOST=test\nTEST_PORT=9090",
		},
	}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := dumpToFile(tt.envs, tt.filename)
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
