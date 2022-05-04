package gazelle

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/rs/zerolog"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/nixconfig"
)

// GenerateRules extracts build metadata from source files in a directory.
// GenerateRules is called in each directory where an update is requested
// in depth-first post-order.
//
// args contains the arguments for GenerateRules. This is passed as a
// struct to avoid breaking implementations in the future when new
// fields are added.
//
// empty is a list of empty rules that may be deleted after merge.
//
// gen is a list of generated rules that may be updated or added.
//
// Any non-fatal errors this function encounters should be logged using
// log.Print.
func (nixLang *nixLang) GenerateRules(
	args language.GenerateArgs,
) language.GenerateResult {
	logger := nixLang.logger.With().
		Str("step", "gazelle.nixLang.GenerateRules").
		Str("path", args.Rel).
		Str("language", languageName).
		Logger()

	logger.Debug().Msg("")
	nixConfigs, ok := args.Config.Exts[languageName].(nixconfig.Configs)

	if !ok {
		logger.Fatal().
			Err(errAssert).
			Msgf("Cannot extract configs")
	}
	cfg, ok := nixConfigs[args.Rel]

	if !ok {
		logger.Fatal().
			Err(errAssert).
			Msgf("Cannot extract config")
	}
	nixPreludeConf := cfg.NixPrelude()

	nixFiles := make(map[string]string)
	nixNames := make(map[string]string)
	nixFilesDeps := make(map[string]*DepSets)

	for _, sourceFile := range append(args.RegularFiles, args.GenFiles...) {
		logger.Trace().
			Str("source", sourceFile).
			Msg("considering")

		if !strings.HasSuffix(sourceFile, "default.nix") {
			continue
		}

		logger.Info().
			Str("file", sourceFile).
			Msg("parsing nix file")

		pth := filepath.Join(args.Dir, sourceFile)
		nixFileDep, err := nixToDepSets(&logger, nixPreludeConf, pth)
		if err != nil {
			continue
		}

		nixFiles[sourceFile] = pth
		nixNames[sourceFile] = args.Rel
		nixFilesDeps[sourceFile] = nixFileDep
	}

	libraries := make(map[string]*nixLibrary)

	for nixFile := range nixFiles {
		fileDeps := nixFilesDeps[nixFile]
		pkgName := strings.ReplaceAll(nixNames[nixFile], "/", ".")

		var tgtName string

		var nixOpts []string

		if len(nixPreludeConf) > 0 {
			tgtName = "//:" + nixPreludeConf
			nixOpts = []string{
				"--argstr",
				"nix_file",
				nixNames[nixFile] + "/default.nix",
			}
		} else {
			tgtName = "//" + nixNames[nixFile] + ":default.nix"
			nixOpts = []string{}
		}

		bzlTemplate := strings.ReplaceAll(
			nixFiles[nixFile],
			"default.nix",
			"BUILD.bazel.tpl",
		)

		var buildFile string

		if fileExists(bzlTemplate) {
			buildFile = "//" + nixNames[nixFile] + ":BUILD.bazel.tpl"
			fileDeps.DepSets[1].Files = append(
				fileDeps.DepSets[1].Files,
				buildFile,
			)
		} else {
			buildFile = ""
		}

		libraries[nixFile] = &nixLibrary{
			Name:      pkgName,
			NixFile:   tgtName,
			Files:     fileDeps.DepSets[1].Files,
			Deps:      fileDeps.DepSets[0].Files,
			NixOpts:   nixOpts,
			BuildFile: buildFile,
		}
	}

	var res language.GenerateResult
	for _, library := range libraries {
		res.Gen = append(res.Gen, library.ToRule())
	}

	res.Imports = make([]interface{}, len(res.Gen))
	for i, r := range res.Gen {
		res.Imports[i] = r.PrivateAttr(config.GazelleImportsKey)
	}

	return res
}

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

// DepSets collection of DepSet structs.
type DepSets struct {
	DepSets []DepSet
}

// DepSet represents dependencies of this package.
type DepSet struct {
	Kind  string
	Files []string
}

// Nix2BuildPath path to a nix evaluator binary.
const Nix2BuildPath = "external/fptrace/bin/fptrace"

func nixToDepSets(logger *zerolog.Logger, nixPrelude, nixFile string) (*DepSets, error) {
	wsroot := os.Getenv("BUILD_WORKSPACE_DIRECTORY")

	scanNix, err := bazel.Runfile(Nix2BuildPath)
	if err != nil {
		logger.Panic().
			Err(err).
			Str("runfile", Nix2BuildPath).
			Msgf("fptrace runfile not found %s", Nix2BuildPath)
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

	depSets := DepSets{[]DepSet{recursive, direct}}

	return &depSets, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

type nixLibrary struct {
	Name      string
	NixFile   string
	BuildFile string
	Files     []string
	Deps      []string
	NixOpts   []string
}

func (l *nixLibrary) ToRule() *rule.Rule {
	ruleStatement := rule.NewRule(exportRule, l.Name)
	ruleStatement.SetAttr("nix_file", l.NixFile)
	sort.Strings(l.Files)
	ruleStatement.SetAttr("files", l.Files)
	ruleStatement.SetAttr("deps", l.Deps)
	ruleStatement.SetAttr("nixopts", l.NixOpts)
	ruleStatement.AddComment("# autogenerated")

	if len(l.BuildFile) > 0 {
		ruleStatement.SetAttr("build_file", l.BuildFile)
	}

	return ruleStatement
}
