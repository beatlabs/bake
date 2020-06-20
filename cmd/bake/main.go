package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"syscall"
	"text/template"
)

// to be overridden by ldflags
var (
	version string = "local"
	commit  string
	date    string
)

const (
	imageName        = "taxibeat/bake"
	dockerSocketFile = "/var/run/docker.sock"
	envVarPrefix     = "BAKE_"
	tpl              = `
		docker run \
		--rm \
		--network {{.NetworkID}} \
		--volume /var/run/docker.sock:/var/run/docker.sock \
		--volume {{.PWD}}:/src \
		--workdir /src \
		--name {{.ContainerName}} \
		--user {{.UserID}}:{{.UserGID}} \
		--group-add {{.DockerSocketGID}} \
		{{- range $k, $v := .EnvVars}}
		--env {{$k}}={{$v}} \
		{{- end}}
		{{.DockerImageName}}:{{.DockerImageTag}} \
		{{.Target}}`
)

type containerArgs struct {
	NetworkID,
	PWD,
	ContainerName,
	UserID,
	UserGID,
	DockerSocketGID,
	DockerImageName,
	DockerImageTag,
	Target string
	EnvVars map[string]string
}

func main() {
	usage := fmt.Sprintf("Available %s commands:\nversion\nrun\n", os.Args[0])
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "version":
		fmt.Printf("Version: %v\nDocker Image Version: %s\nCommit: %s\nDate: %s\n", version, version, commit, date)
		os.Exit(0)
	case "run":
		target := strings.Join(os.Args[2:], " ")
		if err := runBake(target); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}

func runBake(target string) error {
	a, err := newContainerArgs(target)
	if err != nil {
		return err
	}

	c, err := parseTemplate(a)
	if err != nil {
		return err
	}

	defer func() {
		if err := cleanup(a); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println("running command:", c)
	return streamCmd(c)
}

func newContainerArgs(target string) (containerArgs, error) {
	ba := containerArgs{}
	ba.Target = target
	ba.DockerImageName = imageName
	ba.DockerImageTag = version
	ba.PWD = os.Getenv("PWD")

	runID, err := genID()
	if err != nil {
		return containerArgs{}, err
	}
	ba.ContainerName = fmt.Sprintf("bake-%s", runID)

	dockerGID, err := getDockerSocketGID()
	if err != nil {
		return containerArgs{}, err
	}
	ba.DockerSocketGID = dockerGID

	net, err := createNetwork(ba.ContainerName)
	if err != nil {
		return containerArgs{}, err
	}
	ba.NetworkID = net

	u, err := user.Current()
	if err != nil {
		return containerArgs{}, err
	}
	ba.UserID = u.Uid
	ba.UserGID = u.Gid

	ba.EnvVars = map[string]string{
		"NETWORK_ID": net,
		"RUN_ID":     fmt.Sprintf("%s-", ba.ContainerName),
	}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, envVarPrefix) {
			e := strings.TrimPrefix(e, envVarPrefix)
			pair := strings.SplitN(e, "=", 2)
			ba.EnvVars[pair[0]] = pair[1]
		}
	}

	return ba, nil
}

func execCmd(c string) (string, error) {
	cmd := exec.Command("sh", "-c", c)
	bs, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(bs))
	if err != nil {
		return "", fmt.Errorf("exec: %s, output: %s, err: %w", c, out, err)
	}
	return out, nil
}

func streamCmd(c string) error {
	cmd := exec.Command("sh", "-c", c)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func genID() (string, error) {
	l := 3
	buff := make([]byte, l)
	_, err := rand.Read(buff)
	if err != nil {
		return "", err
	}
	str := base64.StdEncoding.EncodeToString(buff)
	return strings.ToLower(str[:l]), nil
}

func getDockerSocketGID() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "0", nil
	case "linux":
		f, err := os.Stat(dockerSocketFile)
		if err != nil {
			return "", err
		}
		gid := fmt.Sprint(f.Sys().(*syscall.Stat_t).Gid)
		return gid, nil

	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func createNetwork(name string) (string, error) {
	out, err := execCmd(fmt.Sprintf("docker network create %s", name))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func cleanup(args containerArgs) error {
	cmds := []string{
		fmt.Sprintf("docker network rm %s", args.NetworkID),
		fmt.Sprintf("docker ps --format '{{.Names}}' | grep %s | awk '{print $1}' | xargs -I {} docker rm -f {}", args.ContainerName),
	}

	for _, c := range cmds {
		_, err := execCmd(c)
		if err != nil {
			return fmt.Errorf("cleanup command: %s: %w", c, err)
		}
	}

	return nil
}

func parseTemplate(args containerArgs) (string, error) {
	t := template.New("bake")
	t, err := t.Parse(tpl)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	if err := t.Execute(&out, args); err != nil {
		return "", err
	}

	return out.String(), nil
}
