package gazelle

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/pathtools"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/rs/zerolog"
)

type (
	TraceOuts []struct {
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
)
type LogEvent struct {
	Path      string
	Runfile   string
	Command   string
	Message   string
	Tracefile string
	Error     error
	Details   []byte
}

func (l *LogEvent) SetMessage(m string) {
	if l.Message == "" {
		l.Message = m
	}
}

func (l LogEvent) Send(logger *zerolog.Logger) {
	fields := map[string]interface{}{
		"command":   l.Command,
		"path":      l.Path,
		"runfile":   l.Runfile,
		"tracefile": l.Tracefile,
		"message":   l.Message,
	}
	for k, v := range fields {
		if v == "" {
			delete(fields, k)
		}
	}

	// Log error
	logger.Error().
		Err(l.Error).
		Fields(fields).
		Send()

	// Log subprocess stderr
	scanner := bufio.NewScanner(bytes.NewReader(l.Details))
	for scanner.Scan() {
		logger.Error().
			Msg(
				fmt.Sprintf(
					"\x1b[%dm%s\x1b[0m",
					31,
					scanner.Text(),
				),
			)
	}

}

func getBazelPackage(workspaceRoot string, filePath string) string {
	var getNixBzlPackage func(string) string
	getNixBzlPackage = func(dir string) string {
		// TODO: NESTED default.nixes in Workspace
		parentDirContainsNixPackageMarker := fileExists(filepath.Join(dir, "default.nix"))
		if parentDirContainsNixPackageMarker {
			return "//" + pathtools.TrimPrefix(
				dir,
				workspaceRoot,
			)
		}

		if dir == "/" {
			panic("Attempted to find the parent directory of '/'")
		}
		return getNixBzlPackage(filepath.Dir(dir))
	}

	fileDir := filepath.Dir(filePath)
	return getNixBzlPackage(fileDir)
}

func getBazelTarget(workspaceRoot string, filePath string) string {
	bazelPackage := getBazelPackage(workspaceRoot, filePath)
	return fmt.Sprintf(
		"%s:%s",
		bazelPackage,
		pathtools.TrimPrefix(
			pathtools.TrimPrefix(filePath, workspaceRoot),
			strings.TrimPrefix(bazelPackage, "//"),
		),
	)
}

func parseFpTraceOutput(workspaceRoot string, rootNixDerivPath string, outputs *TraceOuts) (_, _ []string) {
	if outputs == nil {
		return
	}
	var isInPackage = func(pkg string, target string) bool {
		// TODO: There has to be a better way!
		return pathtools.HasPrefix(
			strings.ReplaceAll(target, ":", string(os.PathSeparator)),
			pkg,
		)
	}
	var filesInRootNixDerivPackage, filesOutsideOfRootNixDerivPackage []string

	rootNixDerivBazelPackage := getBazelPackage(workspaceRoot, rootNixDerivPath)
	for _, output := range *outputs {
		for _, filePath := range output.Inputs {
			// Skip parsing files outside of Bazel workspace
			if !pathtools.HasPrefix(filePath, workspaceRoot) {
				continue
			}

			bazelTarget := getBazelTarget(workspaceRoot, filePath)
			if isInPackage(rootNixDerivBazelPackage, bazelTarget) {
				filesInRootNixDerivPackage = append(filesInRootNixDerivPackage, bazelTarget)
			} else {
				filesOutsideOfRootNixDerivPackage = append(filesOutsideOfRootNixDerivPackage, bazelTarget)
			}
		}
	}

	return filesInRootNixDerivPackage, filesOutsideOfRootNixDerivPackage
}

func nixToDepSets(logger *zerolog.Logger, nixPrelude, nixFile string) (_, _ []string, err error) {
	// TODO: Lookupenv
	wsroot := os.Getenv("BUILD_WORKSPACE_DIRECTORY")

	le := &LogEvent{
		Path:    nixFile,
		Runfile: FPTRACE_PATH,
	}

	defer err2.Handle(&err, func() {
		le.Error = err
		le.Send(logger)
	})

	// TODO: distinguish between fatal/non fatal errors
	pathToFpTrace := try.To1(bazel.Runfile(FPTRACE_PATH))
	tmpfile := try.To1(ioutil.TempFile("", "nix-gzl*.json"))

	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	nixInstantiatePreludeParams := ""
	if len(nixPrelude) > 0 {
		// Thank you golang, that I may not name string formating params, that makes things so much more readable
		// Should use: os.PathSeparator
		// <workspace-root-path>/<nix-prelude-file> --args nixfile=
		nixInstantiatePreludeParams = fmt.Sprintf("%s/%s --argstr nix_file ", wsroot, nixPrelude)
	}

	// Thank you golang, that I may not name string formating params, that makes things so much more readable
	// -d <tmp-dir-path> nix-instantiate <nixPrelude><path-to-nix-file>
	var fpTraceParams = fmt.Sprintf(
		"-d %s nix-instantiate %s%s",
		tmpfile.Name(),
		nixInstantiatePreludeParams,
		nixFile,
	)

	var outputBuf bytes.Buffer
	cmd := exec.Command(
		pathToFpTrace,
		strings.Split(fpTraceParams, " ")...,
	)
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	defer err2.Handle(&err, func() {
		le.Details = outputBuf.Bytes()
		le.Command = strings.Join(cmd.Args, " ")
		le.SetMessage("evaluation of nix expression failed")
	})
	try.To(cmd.Run())

	defer err2.Handle(&err, func() {
		le.Tracefile = tmpfile.Name()
		le.SetMessage("unmarshaling of trace output failed")
	})
	byteValue := try.To1(os.ReadFile(tmpfile.Name()))

	var traceOuts TraceOuts
	try.To(json.Unmarshal(byteValue, &traceOuts))

	directDeps, externalDeps := parseFpTraceOutput(wsroot, nixFile, &traceOuts)
	return directDeps, externalDeps, nil
}
