package gazelle

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/rs/zerolog"
)

type TraceOut struct {
	Cmd struct {
		Parent int
		ID     int
		Dir    string
		Path   string
		Args   []string
	}
	Inputs  []string
	Outputs []string
	FDs     struct {
		Num0 string
		Num1 string
		Num2 string
	}
}

// DepSet represents dependencies of this package.
type DepSet struct {
	Kind  string
	Files []string
}

func nixToDepSets(logger *zerolog.Logger, nixPrelude, nixFile string) ([]DepSet, error) {
	wsroot := os.Getenv("BUILD_WORKSPACE_DIRECTORY")

	scanNix, err := bazel.Runfile(nix2BuildPath)
	if err != nil {
		logger.Panic().
			Err(err).
			Str("runfile", nix2BuildPath).
			Msgf("fptrace runfile not found %s", nix2BuildPath)
	}

	tmpfile, err := ioutil.TempFile("", "nix-gzl*.json")
	if err != nil {
		logger.Panic().
			Err(err).
			Msgf("could not create fptrace output file")
	}

	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	var cmd *exec.Cmd

	if len(nixPrelude) > 0 {
		cmd = exec.Command(
			scanNix,
			"-d",
			tmpfile.Name(),
			"nix-instantiate",
			wsroot+"/"+nixPrelude,
			"--argstr",
			"nix_file",
			nixFile,
		)
	} else {
		cmd = exec.Command(scanNix, "-d", tmpfile.Name(), "nix-instantiate", nixFile)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		details := strings.Split(string(out[:]), "\n")
		details = details[:len(details)-1]
		logger.Error().
			Err(err).
			Str("path", nixFile).
			Msg("evaluation of nix expression failed")

		for i := range details {
			logger.Error().Msg(details[i])
		}

		return nil, err
	}

	var traceOuts []TraceOut

	byteValue, _ := os.ReadFile(tmpfile.Name())
	err = json.Unmarshal(byteValue, &traceOuts)

	if err != nil {
		logger.Error().
			Err(err).
			Str("path", nixFile).
			Msg("unmarshaling of trace output failed")
		return nil, err
	}

	filteredFiles := []string{nixFile}

	var traceOut TraceOut
	for i := range traceOuts {
		traceOut = traceOuts[i]
		for j := range traceOut.Inputs {
			considered := traceOut.Inputs[j]
			if considered != nixFile && strings.HasPrefix(considered, wsroot) {
				filteredFiles = append(filteredFiles, considered)
			}
		}
	}

	sort.Strings(filteredFiles)
	sort.Slice(filteredFiles, func(i, j int) bool {
		return len(filteredFiles[i]) > len(filteredFiles[j])
	})

	packages := []string{}

	for i := range filteredFiles {
		considered := filteredFiles[i]
		if strings.HasSuffix(considered, "default.nix") {
			packages = append(
				packages,
				strings.TrimSuffix(considered, "default.nix"),
			)
		}
	}

	direct := DepSet{"direct", []string{}}
	recursive := DepSet{"recursive", []string{}}
	targets := []string{}

	for _, consideredPackage := range packages {
		temp := filteredFiles[:0]

		for _, consideredFile := range filteredFiles {
			if strings.HasPrefix(consideredFile, consideredPackage) {
				pkg := strings.TrimSuffix(
					strings.TrimPrefix(consideredPackage, wsroot+"/"),
					"/",
				)
				reltarget := strings.TrimPrefix(consideredFile, consideredPackage)
				target := "//" + pkg + ":" + reltarget
				targets = append(targets, target)
			} else {
				temp = append(temp, consideredFile)
			}
		}

		filteredFiles = temp
	}

	nixPackage := "//" + strings.TrimPrefix(
		strings.TrimSuffix(nixFile, "/default.nix"),
		wsroot+"/",
	)
	for _, x := range targets {
		if strings.HasPrefix(x, nixPackage) {
			direct.Files = append(direct.Files, x)
		}

		recursive.Files = append(recursive.Files, x)
	}

	return []DepSet{recursive, direct}, nil
}
