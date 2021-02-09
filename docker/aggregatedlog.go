package docker

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type colorFunc func(string, ...interface{}) string

var colors = []colorFunc{
	color.BlueString,
	color.CyanString,
	color.GreenString,
	color.MagentaString,
	color.RedString,
	color.YellowString,
}

func streamContainerLogs(name string, cf colorFunc) {
	pool, _ := dockertest.NewPool("")
	w := &prefixedWriter{target: os.Stdout, prefix: cf(name + " | ")}
	err := pool.Client.Logs(
		docker.LogsOptions{
			Container:    name,
			OutputStream: w,
			ErrorStream:  w,
			Follow:       true,
			Stdout:       true,
			Stderr:       true,
		},
	)
	if err != nil {
		fmt.Printf("Could not attach logs to %s: %v\n", name, err)
	}
}

type prefixedWriter struct {
	target io.Writer
	prefix string
}

func (w *prefixedWriter) Write(b []byte) (int, error) {
	fmt.Fprintf(w.target, w.prefix)
	return w.target.Write(b)
}
