package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mage"
	"github.com/magefile/mage/sh"
)

const bakeStaticBinary = "./bake-static"

func main() {
	forceStaticBinary := false
	if os.Getenv("BAKE_FORCE_STATIC_BINARY") != "" {
		forceStaticBinary = true
	}

	_, err := exec.LookPath("go")
	if err != nil || forceStaticBinary {
		staticBinary()
		os.Exit(0)
	}

	os.Exit(mage.Main())
}

func staticBinary() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v\n", err)
	}

	args := []string{
		"run", "--rm", "--volume", wd + ":/app", "--workdir", "/app", "golang:1.14", "sh", "-c",
		"git config --global url.\"https://golang:$GITHUB_TOKEN@github.com\".insteadOf \"https://github.com\" && " +
			"go get github.com/magefile/mage@v1.10.0 && " + "mage -compile " + bakeStaticBinary + " && " +
			"chown --reference=./ " + bakeStaticBinary,
	}
	fmt.Printf("Executing cmd: %s\n", strings.Join(args, " "))

	err = sh.RunV("docker", args...)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Executing cmd: %s %s\n", bakeStaticBinary, strings.Join(os.Args[1:], " "))
	err = sh.RunV(bakeStaticBinary, os.Args[1:]...)
	if err != nil {
		log.Fatal(err)
	}
}
