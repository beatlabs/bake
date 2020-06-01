package github

import (
	"errors"
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
)

const (
	git = "git"
)

func SetupToken() error {
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		return errors.New("env var GH_TOKEN for the GitHub access token is not set")
	}
	if token == "" {
		return errors.New("access token is empty")
	}

	args := []string{
		"config",
		fmt.Sprintf(`url.”https://%s:@github.com/".insteadOf`, token),
		`“https://github.com/"`,
	}

	return sh.RunV(git, args...)
}
