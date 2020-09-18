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

func main() {

	_, err := exec.LookPath("go")
	if err == nil {
		os.Exit(mage.Main())
	}

	// go is not installed, let's use docker to build a static, project specific, bake binary
	cmd := "docker"
	bake := "./bake-static"

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v\n", err)
	}

	args := []string{"run", "--rm", "--volume", wd + ":/app", "--workdir", "/app", "golang:1.14", "sh", "-c",
		"go get github.com/magefile/mage@v1.9.0 && " +
			"mage -compile " + bake + " && " +
			"chown --reference=./ " + bake + " "}
	fmt.Printf("Executing cmd: %s %s\n", cmd, strings.Join(args, " "))

	err = sh.RunV(cmd, args...)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Executing cmd: %s %s\n", bake, strings.Join(args, " "))
	err = sh.RunV(bake, os.Args[1:]...)
	if err != nil {
		log.Fatal(err)
	}
}
