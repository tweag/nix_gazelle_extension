package gazelle

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
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
	nixRepositoriesConf := cfg.NixRepositories()

	var res language.GenerateResult

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

		//TODO: parser should be launched only when generating
		//workspace rules
		nixFileDep, err := nixToDepSets(&logger, nixPreludeConf, pth)
		if err != nil {
			continue
		}

		pkgName := strings.ReplaceAll(args.Rel, "/", ".")

		var tgtName string

		var nixOpts []string

		if len(nixPreludeConf) > 0 {
			tgtName = "//:" + nixPreludeConf
			nixOpts = []string{
				"--argstr",
				"nix_file",
				args.Rel + "/default.nix",
			}
		} else {
			tgtName = "//" + args.Rel + ":default.nix"
			nixOpts = []string{}
		}

		// TODO: instead of using template file
		// use already existing/generated one.
		bzlTemplate := strings.ReplaceAll(
			sourceFile,
			"default.nix",
			"BUILD.bazel.tpl",
		)

		var buildFile string

		if fileExists(bzlTemplate) {
			buildFile = "//" + args.Rel + ":BUILD.bazel.tpl"
			nixFileDep.DepSets[1].Files = append(
				nixFileDep.DepSets[1].Files,
				buildFile,
			)
		} else {
			buildFile = ""
		}

		if cfg.WsMode() {
			res.Gen = append(res.Gen, (&nixPackage{
				name:         pkgName,
				nixFile:      tgtName,
				nixFileDeps:  nixFileDep.DepSets[0].Files,
				nixOpts:      nixOpts,
				buildFile:    buildFile,
				repositories: nixRepositoriesConf,
			}).ToRule())
		} else {
			res.Gen = append(res.Gen, (&nixExport{
				name:  pkgName,
				files: nixFileDep.DepSets[1].Files,
			}).ToRule())
		}
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
