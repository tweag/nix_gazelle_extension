package gazelle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/nixconfig"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/private/logconfig"
)

func SourceFileToNixRules(
	sourceFile string,
	sourceDirAbs string,
	sourceDirRel string,
	nixCfg *nixconfig.NixLanguageConfig,
	wg *sync.WaitGroup,
	rules chan<- rule.Rule) {
	defer wg.Done()

	var logger = logconfig.GetLogger()
	logger.Trace().
		Str("source", sourceFile).
		Msg("considering")

	if sourceFile != "default.nix" {
		return
	}

	logger.Info().
		Str("file", sourceFile).
		Msg("parsing nix file")

	pth := filepath.Join(sourceDirAbs, sourceFile)

	directDeps, chainedDeps, err := nixToDepSets(logger, nixCfg.NixPrelude, pth)
	if err != nil {
		return
	}

	pkgName := strings.ReplaceAll(sourceDirRel, "/", ".")

	// TODO: instead of using template file
	// use already existing/generated one.
	bzlTemplate := strings.ReplaceAll(
		sourceFile,
		"default.nix",
		"BUILD.bazel.tpl",
	)

	buildFile := fmt.Sprintf("//%s:BUILD.bazel.tpl", sourceDirRel)

	nrap := &NixRuleArgs{
		kind: MANIFEST_RULE,
		attrs: map[string]interface{}{
			"name":          pkgName,
			"nix_file_deps": chainedDeps,
			"repositories":  nixCfg.NixRepositories,
		},
		comments: []string{
			"# autogenerated",
		},
	}

	if len(nixCfg.NixPrelude) > 0 {
		nrap.attrs["nix_file"] = fmt.Sprintf("//:%s", nixCfg.NixPrelude)
		nrap.attrs["nix_opts"] = []string{
			"--argstr",
			"nix_file",
			filepath.Join(sourceDirRel, sourceFile),
		}
	} else {
		nrap.attrs["nix_file"] = fmt.Sprintf("//%s:%s", sourceDirRel, sourceFile)
	}

	rules <- *genNixRule(nrap)

	if fileExists(bzlTemplate) {
		nrap.attrs["build_file"] = buildFile
		directDeps = append(
			directDeps,
			buildFile,
		)
	}

	nrae := &NixRuleArgs{
		kind: EXPORT_RULE,
		attrs: map[string]interface{}{
			"name":  "exports",
			"files": directDeps,
		},
		comments: []string{
			"# autogenerated",
		},
	}

	rules <- *genNixRule(nrae)
}

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
		Str("language", LANGUAGE_NAME).
		Logger()

	logger.Debug().Msg("")
	nixConfigs, ok := args.Config.Exts[LANGUAGE_NAME].(nixconfig.NixLanguageConfigs)
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

	var res language.GenerateResult
	var wg sync.WaitGroup
	var rules = make(chan rule.Rule)

	go func() {
		wg.Wait()
		close(rules)
	}()

	for _, sourceFile := range append(args.RegularFiles, args.GenFiles...) {
		wg.Add(1)
		go SourceFileToNixRules(sourceFile, args.Dir, args.Rel, cfg, &wg, rules)
	}

	// Read of channel is blocking
	for r := range rules {
		res.Gen = append(res.Gen, &r)
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
