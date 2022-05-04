package gazelle

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
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

	// TODO: use structs for metadata. i.e. nixTarget
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

	libraries := make(map[string]*nixPackage)

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

		libraries[nixFile] = &nixPackage{
			name:        pkgName,
			nixFile:     tgtName,
			files:       fileDeps.DepSets[1].Files,
			nixFileDeps: fileDeps.DepSets[0].Files,
			nixOpts:     nixOpts,
			buildFile:   buildFile,
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func (pkg *nixPackage) ToRule() *rule.Rule {
	ruleStatement := rule.NewRule(exportRule, pkg.name)
	ruleStatement.SetAttr("nix_file", pkg.nixFile)
	sort.Strings(pkg.files)
	ruleStatement.SetAttr("files", pkg.files)
	ruleStatement.SetAttr("nix_file_deps", pkg.nixFileDeps)
	ruleStatement.SetAttr("nixopts", pkg.nixOpts)
	ruleStatement.SetAttr("repositories", map[string]string{"": ""}) //TODO: actually set based on directive contents
	ruleStatement.AddComment("# autogenerated")

	if len(pkg.buildFile) > 0 {
		ruleStatement.SetAttr("build_file", pkg.buildFile)
	}

	return ruleStatement
}
