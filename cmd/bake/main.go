package main

import (
	"context"
	"errors"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/ory/dockertest/v3/docker/pkg/stdcopy"
)

const (
	imageName = "ghcr.io/taxibeat/bake"
	imageTag  = "latest"
)

func main() {
	// TODO: need to provide arg or env var
	skipCleanup := false

	dockerSock, dockerGID, err := getDockerData()
	if err != nil {
		log.Panicln(err)
	}

	sessionID := getSessionID()

	cli, err := createDockerClient()
	if err != nil {
		log.Panicln(err)
	}

	// TODO: arg or env var?
	timeout := 1 * time.Minute

	ctx, cnl := context.WithTimeout(context.Background(), timeout)
	defer cnl()

	networkID, err := createDockerNetwork(ctx, cli, sessionID)
	if err != nil {
		log.Fatalln(err)
	}
	defer removeDockerNetwork(ctx, cli, networkID, skipCleanup)

	containerID, err := dockerRun(ctx, cli, networkID)
	if err != nil {
		log.Fatalln(err)
	}
	defer containerCleanup(ctx, cli, containerID, skipCleanup)
}

func getDockerData() (sock string, gid int, err error) {
	msg := "docker details, sock: %s gid: %d"
	if strings.HasPrefix(runtime.GOOS, "linux") {
		sock = "/var/run/docker.sock"

		var stat syscall.Stat_t

		err = syscall.Stat(sock, &stat)
		if err != nil {
			return
		}
		gid = int(stat.Gid)
		log.Printf(msg, sock, gid)
		return
	}

	if strings.HasPrefix(runtime.GOOS, "darwin") {

		sock = "/var/run/docker.sock.raw"
		gid = 0
		log.Printf(msg, sock, gid)
		return
	}
	err = errors.New("unsupported os")
	return
}

func getSessionID() string {
	sessionID := uuid.New().String()
	sessionID = sessionID[:3]
	sessionID = strings.ToLower(strings.ToUpper(sessionID))
	log.Printf("session id: %s", sessionID)
	return sessionID
}

func createDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv)
}

func createDockerNetwork(ctx context.Context, cli *client.Client, sessionID string) (string, error) {
	// TODO: do i need to provide options?
	options := types.NetworkCreate{}

	rsp, err := cli.NetworkCreate(ctx, sessionID, options)
	if err != nil {
		return "", err
	}
	if rsp.Warning != "" {
		log.Printf("warning creating network: %s", rsp.Warning)
	}
	return rsp.ID, nil
}

func removeDockerNetwork(ctx context.Context, cli *client.Client, networkID string, skip bool) {
	if skip {
		log.Println("skipping docker network cleanup")
		return
	}
	err := cli.NetworkRemove(ctx, networkID)
	if err != nil {
		log.Printf("failed to remove docker network with id %s: %v", networkID, err)
	}
}

func dockerRun(ctx context.Context, cli *client.Client, networkID string) (string, error) {
	config := &container.Config{
		Hostname:        "",
		Domainname:      "",
		User:            "",
		AttachStdin:     false,
		AttachStdout:    false,
		AttachStderr:    false,
		ExposedPorts:    nil,
		Tty:             false,
		OpenStdin:       false,
		StdinOnce:       false,
		Env:             nil,
		Cmd:             nil,
		Healthcheck:     nil,
		ArgsEscaped:     false,
		Image:           "",
		Volumes:         nil,
		WorkingDir:      "",
		Entrypoint:      nil,
		NetworkDisabled: false,
		MacAddress:      "",
		OnBuild:         nil,
		Labels:          nil,
		StopSignal:      "",
		StopTimeout:     nil,
		Shell:           nil,
	}
	hostConfig := &container.HostConfig{
		Binds:           nil,
		ContainerIDFile: "",
		LogConfig:       container.LogConfig{},
		NetworkMode:     "",
		PortBindings:    nil,
		RestartPolicy:   container.RestartPolicy{},
		AutoRemove:      true,
		VolumeDriver:    "",
		VolumesFrom:     nil,
		CapAdd:          nil,
		CapDrop:         nil,
		CgroupnsMode:    "",
		DNS:             nil,
		DNSOptions:      nil,
		DNSSearch:       nil,
		ExtraHosts:      nil,
		GroupAdd:        nil,
		IpcMode:         "",
		Cgroup:          "",
		Links:           nil,
		OomScoreAdj:     0,
		PidMode:         "",
		Privileged:      false,
		PublishAllPorts: false,
		ReadonlyRootfs:  false,
		SecurityOpt:     nil,
		StorageOpt:      nil,
		Tmpfs:           nil,
		UTSMode:         "",
		UsernsMode:      "",
		ShmSize:         0,
		Sysctls:         nil,
		Runtime:         "",
		ConsoleSize:     [2]uint{},
		Isolation:       "",
		Resources:       container.Resources{},
		Mounts:          nil,
		MaskedPaths:     nil,
		ReadonlyPaths:   nil,
		Init:            nil,
	}
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: nil,
	}
	platform := &specs.Platform{
		Architecture: "",
		OS:           "",
		OSVersion:    "",
		OSFeatures:   nil,
		Variant:      "",
	}

	containerName := "string"

	rsp, err := cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
	if err != nil {
		return "", err
	}

	log.Printf("container %s created. warnings %s", rsp.ID, strings.Join(rsp.Warnings, ","))

	err = cli.ContainerStart(ctx, rsp.ID, types.ContainerStartOptions{})
	if err != nil {
		return rsp.ID, err
	}

	err = cli.NetworkConnect(ctx, networkID, rsp.ID, nil)
	if err != nil {
		return rsp.ID, err
	}

	statusCh, errCh := cli.ContainerWait(ctx, rsp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return rsp.ID, err
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, rsp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return rsp.ID, err
	}

	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	if err != nil {
		log.Printf("failed to copy docker std out: %v", err)
	}

	return rsp.ID, nil
}

func containerCleanup(ctx context.Context, cli *client.Client, id string, skip bool) {
	if skip {
		log.Println("skipping docker container cleanup")
		return
	}
	// TODO: static or configurable?
	timeout := 10 * time.Second
	err := cli.ContainerStop(ctx, id, &timeout)
	if err != nil {
		log.Printf("failed to stop container with id %s", id)
	}

	err = cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{RemoveVolumes: true, RemoveLinks: true, Force: true})
	if err != nil {
		log.Printf("failed to remove container with id %s", id)
	}
}
