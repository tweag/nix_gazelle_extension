package gazelle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/pathtools"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/rs/zerolog"
)

type TraceOuts []struct {
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

func nixToDepSets(logger *zerolog.Logger, nixPrelude, nixFile string) (_, _ []string, err error) {
	wsroot := os.Getenv("BUILD_WORKSPACE_DIRECTORY")

	errfields := make(map[string]interface{}, 0)
	out := make([]byte, 0)

	defer err2.Handle(&err, func() {
		logger.Error().Fields(errfields).Send()
		errdetails := strings.Split(string(out[:]), "\n")
		errdetails = errdetails[:len(errdetails)-1]
		for _, ed := range errdetails {
			logger.Error().Msg(fmt.Sprintf("\x1b[%dm%s\x1b[0m", 31, ed))
		}
	})

	errfields["path"] = nixFile
	errfields["runfile"] = NIX2BUILDPATH
	scanNix := try.To1(bazel.Runfile(NIX2BUILDPATH))

	tmpfile := try.To1(ioutil.TempFile("", "nix-gzl*.json"))

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

	out, err = cmd.CombinedOutput()
	defer err2.Handle(&err, func() {
		errfields["cmd"] = strings.Join(cmd.Args, " ")
		if _, k := errfields["message"]; !k {
			errfields["message"] = "evaluation of nix expression failed"
		}
	})
	if err != nil {
		return nil, nil, err
	}

	defer err2.Handle(&err, func() {
		errfields["traceout"] = tmpfile.Name()
		if _, k := errfields["message"]; !k {
			errfields["message"] = "unmarshaling of trace output failed"
		}
	})

	byteValue := try.To1(os.ReadFile(tmpfile.Name()))

	var traceOuts TraceOuts
	try.To(json.Unmarshal(byteValue, &traceOuts))

	filteredFiles := []string{nixFile}

	for _, traceOut := range traceOuts {
		for _, considered := range traceOut.Inputs {
			if considered != nixFile && pathtools.HasPrefix(considered, wsroot) {
				filteredFiles = append(filteredFiles, considered)
			}
		}
	}

	sort.Strings(filteredFiles)
	sort.Slice(filteredFiles, func(i, j int) bool {
		return len(filteredFiles[i]) > len(filteredFiles[j])
	})

	packages := []string{}

	for _, considered := range filteredFiles {
		if strings.HasSuffix(considered, "default.nix") {
			packages = append(
				packages,
				strings.TrimSuffix(considered, "default.nix"),
			)
		}
	}

	directDeps := []string{}
	chainedDeps := []string{}
	targets := []string{}

	for _, consideredPackage := range packages {
		temp := filteredFiles[:0]

		for _, consideredFile := range filteredFiles {
			if strings.HasPrefix(consideredFile, consideredPackage) {
				pkg := trimTrailingSlash(pathtools.TrimPrefix(consideredPackage, wsroot))
				reltarget := pathtools.TrimPrefix(consideredFile, consideredPackage)
				target := fmt.Sprintf("//%s:%s", pkg, reltarget)
				targets = append(targets, target)
			} else {
				temp = append(temp, consideredFile)
			}
		}

		filteredFiles = temp
	}

	nixPackage := "//" + pathtools.TrimPrefix(
		strings.TrimSuffix(nixFile, "/default.nix"),
		wsroot,
	)
	for _, x := range targets {
		if strings.HasPrefix(x, nixPackage) {
			directDeps = append(directDeps, x)
		}

		chainedDeps = append(chainedDeps, x)
	}

	return directDeps, chainedDeps, nil
}

func trimTrailingSlash(p string) string {
	for len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	return p
}
