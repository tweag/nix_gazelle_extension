package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"os"
	"os/exec"
)

type DepSets struct {
	DepSets []DepSet
}

type DepSet struct {
	Kind  string
	Files []string
}

const NIX2BUILD_PATH = "external/scan_nix/bin/scan-nix"

func nixToDepSets(nixFile string) DepSets {
	scan_nix, err := bazel.Runfile(NIX2BUILD_PATH)
	cmd := exec.Command(scan_nix, nixFile)
	cmd.Dir = os.Getenv("BUILD_WORKSPACE_DIRECTORY")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s", out)
		log.Fatal(err)
	}
	var depSets DepSets
	err = json.Unmarshal(out, &depSets)
	if err != nil {
		log.Printf("Incorrect json: %s\n", out)
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", depSets)
	return depSets
}

func main() {
	nixToDepSets(os.Getenv("BUILD_WORKSPACE_DIRECTORY") + "/examples/folks/cool-kid/default.nix")
}

