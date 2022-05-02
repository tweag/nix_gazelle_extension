package gazelle

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tweag/nix_gazelle_extension/nix/gazelle/private/logconfig"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	golang "github.com/bazelbuild/bazel-gazelle/language/go"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/bazel-gazelle/walk"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/rs/zerolog"
)

const (
	nixName     = "nix"
	exportRule  = "export_nix"
	packageRule = "nixpkgs_package"
)

var (
	errAssert = errors.New("assertion failed")
	errParse  = errors.New("directive parsing failed")
)

type nixLang struct{
	logger zerolog.Logger
}

// NewLanguage implementation.
func NewLanguage() language.Language {

	logLevel := logconfig.LogLevel()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(logLevel)
	logger.Debug().Msg("creating nix language")

	nl := nixLang{
		logger: logger,
	}

	return &nl
}

func (l *nixLang) Name() string { return nixName }

// Kinds returns a map of maps rule names (kinds) and information on how to
// match and merge attributes that may be found in rules of those kinds. All
// kinds of rules generated for this language may be found here.
func (l *nixLang) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		exportRule: {
			MatchAny:   false,
			MatchAttrs: []string{"name"},
		},
		packageRule: {
			MatchAttrs: []string{"name", "nix_file_deps"},
			MergeableAttrs: map[string]bool{
				"nix_file_deps": true,
			},
		},
	}
}

// Loads returns .bzl files and symbols they define. Every rule generated by
// GenerateRules, now or in the past, should be loadable from one of these
// files.
func (l *nixLang) Loads() []rule.LoadInfo {
	return []rule.LoadInfo{
		{
			Name: "@io_tweag_gazelle_nix//nix:defs.bzl",
			Symbols: []string{
				exportRule,
			},
		},
		{
			Name: "@io_tweag_rules_nixpkgs//nixpkgs:nixpkgs.bzl",
			Symbols: []string{
				"nixpkgs_package",
			},
		},
	}
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
func (l *nixLang) GenerateRules(
	args language.GenerateArgs,
) language.GenerateResult {
	nixConfig, ok := args.Config.Exts[nixName].(Config)

	if !ok {
		l.logger.Fatal().
			Err(errAssert).
			Str("config", nixName).
			Msgf("Cannot extract directive %s", nixName)
	}

	nixPreludeConf := nixConfig.NixPrelude

	nixFiles := make(map[string]string)
	nixNames := make(map[string]string)
	nixFilesDeps := make(map[string]*DepSets)

	for _, sourceFile := range append(args.RegularFiles, args.GenFiles...) {
		if !strings.HasSuffix(sourceFile, "default.nix") {
			continue
		}

		pth := filepath.Join(args.Dir, sourceFile)
		l.logger.Info().
			Str("path", pth).
			Msg("parsing nix file")

		nixFileDep, err := nixToDepSets(l.logger, nixPreludeConf, pth)
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

func fixGazelle(buildFile *rule.File) {
	for _, loadStatement := range buildFile.Loads {
		if loadStatement.Has(exportRule) {
			loadStatement.Remove(exportRule)

			if loadStatement.IsEmpty() {
				loadStatement.Delete()
			}
		}
	}

	var knownRuleStatements []*rule.Rule

	for _, ruleStatement := range buildFile.Rules {
		if ruleStatement.Kind() == exportRule {
			knownRuleStatements = append(knownRuleStatements, ruleStatement)
		}
	}

	for _, knownRuleStatement := range knownRuleStatements {
		knownRuleStatement.Delete()
	}
}

// Fix repairs deprecated usage of language-specific rules in f. This is
// called before the file is indexed. Unless c.ShouldFix is true, fixes
// that delete or rename rules should not be performed.
func (l *nixLang) Fix(c *config.Config, f *rule.File) {
	fixGazelle(f)
}

func (*nixLang) CheckFlags(
	fs *flag.FlagSet,
	c *config.Config,
) error {
	return nil
}

// Config configuration for language extension.
type Config struct {
	NixPrelude      string
	NixRepositories map[string]string
}

// Configure modifies the configuration using directives and other
// information extracted from a build file. Configure is called in
// each directory.
//
// c is the configuration for the current directory. It starts out as
// a copy of the configuration for the parent directory.
//
// rel is the slash-separated relative path from the repository root
// to the current directory. It is "" for the root directory itself.
//
// f is the build file for the current directory or nil if there is no
// existing build file.

func (l *nixLang) Configure(extensionConfig *config.Config, rel string, buildFile *rule.File) {
	if buildFile == nil {
		return
	}

	var extraConfig Config

	m, ok := extensionConfig.Exts[nixName].(Config)

	if ok {
		extraConfig = m
	} else {
		extraConfig = Config{
			NixPrelude:      "",
			NixRepositories: make(map[string]string),
		}
	}

	for _, directive := range buildFile.Directives {
		switch directive.Key {
		case "nix_prelude":
			extraConfig.NixPrelude = directive.Value
		case "nix_repositories":
			parseNixRepositories(l.logger, &extraConfig, directive.Value)
		}
	}

	extensionConfig.Exts[nixName] = extraConfig
}

func parseNixRepositories(logger zerolog.Logger, nixconfig *Config, value string) {
	const parts = 2

	pairs := strings.Split(value, " ")
	keyValuePair := make([]string, 0, len(pairs)*parts)

	for _, key := range pairs {
		keyValuePair = append(keyValuePair, strings.Split(key, "=")...)
	}

	directive := "nix_repositories"
	if len(keyValuePair)%2 != 0 {
		logger.Panic().
			Err(errParse).
			Str("value", value).
			Str("directive", directive).
			Msgf("Cannot parse %s directive, invalid value %s", directive, value)
	}

	for i := 0; i < len(keyValuePair); i += 2 {
		nixconfig.NixRepositories[keyValuePair[i]] = keyValuePair[i+1]
	}
}

// Embeds returns a list of labels of rules that the given rule
// embeds. If a rule is embedded by another importable rule of the
// same language, only the embedding rule will be indexed. The
// embedding rule will inherit the imports of the embedded rule.
func (l *nixLang) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

// Imports returns a list of ImportSpecs that can be used to import
// the rule r. This is used to populate RuleIndex.
//
// If nil is returned, the rule will not be indexed. If any non-nil
// slice is returned, including an empty slice, the rule will be
// indexed.
func (l *nixLang) Imports(
	extensionConfig *config.Config,
	ruleStatement *rule.Rule,
	buildFile *rule.File,
) []resolve.ImportSpec {
	var prefix string

	switch ruleStatement.Kind() {
	case exportRule:
		prefix = "exports:"
	case packageRule:
		prefix = "nixpkgs_package:"
	}

	return []resolve.ImportSpec{{nixName, prefix + ruleStatement.Name()}}
}

// KnownDirectives returns a list of directive keys that this
// Configurer can interpret. Gazelle prints errors for directives that
// are not recoginized by any Configurer.
func (*nixLang) KnownDirectives() []string {
	return []string{
		"nix_prelude",
		"nix_repositories",
	}
}

// RegisterFlags registers command-line flags used by the
// extension. This method is called once with the root configuration
// when Gazelle starts. RegisterFlags may set an initial values in
// Config.Exts. When flags are set, they should modify these values.
func (l *nixLang) RegisterFlags(
	fs *flag.FlagSet,
	cmd string,
	c *config.Config,
) {
}

// Resolve translates imported libraries for a given rule into Bazel
// dependencies. A list of imported libraries is typically stored in a
// private attribute of the rule when it's generated (this interface
// doesn't dictate how that is stored or represented). Resolve
// generates a "deps" attribute (or the appropriate language-specific
// equivalent) for each import according to language-specific rules
// and heuristics.
func (l *nixLang) Resolve(
	c *config.Config,
	ix *resolve.RuleIndex,
	rc *repo.RemoteCache,
	r *rule.Rule,
	importsRaw interface{},
	from label.Label,
) {
}

type nixWorkspaceLibrary struct {
	Name         string
	NixFile      string
	NixFileDeps  []string
	Repositories map[string]string
	NixOpts      []string
	BuildFile    string
}

func (l *nixLang) UpdateRepos(
	args language.UpdateReposArgs,
) language.UpdateReposResult {
	packageList := collectDependenciesFromRepo(args.Config, l)
	nixConfig, ok := args.Config.Exts[nixName].(Config)

	if !ok {
		return language.UpdateReposResult{
			Error: fmt.Errorf("%w %s", errAssert, "nixConfig"),
			Gen:   nil,
		}
	}

	nixRepositoriesConf := nixConfig.NixRepositories

	repoRuleStatements := make([]*rule.Rule, len(packageList))

	for idx, pkg := range packageList {
		ruleStatement := rule.NewRule("nixpkgs_package", pkg.Name)
		ruleStatement.SetAttr("nix_file", pkg.NixFile)
		ruleStatement.SetAttr("nixopts", pkg.NixOpts)
		ruleStatement.SetAttr("nix_file_deps", pkg.NixFileDeps)
		ruleStatement.SetAttr("repositories", nixRepositoriesConf)

		if len(pkg.BuildFile) > 0 {
			ruleStatement.SetAttr("build_file", pkg.BuildFile)
		}

		ruleStatement.AddComment("# autogenerated")
		repoRuleStatements[idx] = ruleStatement
	}

	return language.UpdateReposResult{
		Error: nil,
		Gen:   repoRuleStatements,
	}
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

func nixToDepSets(logger zerolog.Logger, nixPrelude, nixFile string) (*DepSets, error) {
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

func collectDependenciesFromRepo(
	extensionConfig *config.Config,
	lang language.Language,
) []nixWorkspaceLibrary {
	packages := make([]nixWorkspaceLibrary, 0)
	cexts := []config.Configurer{
		&config.CommonConfigurer{},
		&walk.Configurer{},
		lang,
		golang.NewLanguage(),
		proto.NewLanguage(),
	}

	initUpdateReposConfig(extensionConfig, cexts)

	walk.Walk(
		extensionConfig,
		cexts,
		[]string{},
		walk.VisitAllUpdateDirsMode,
		func(
			dir,
			rel string,
			extensionConfig *config.Config,
			update bool,
			buildFile *rule.File,
			subdirs,
			regularFiles,
			genFiles []string,
		) {
			collectDependenciesFromFile(buildFile, &packages)
		},
	)

	return packages
}

func collectDependenciesFromFile(
	buildFile *rule.File,
	packages *[]nixWorkspaceLibrary,
) {
	if buildFile == nil {
		return
	}

	for _, ruleStatement := range buildFile.Rules {
		if ruleStatement.Kind() == exportRule {
			pkg := nixWorkspaceLibrary{
				Name:         ruleStatement.AttrString("name"),
				NixFile:      ruleStatement.AttrString("nix_file"),
				NixFileDeps:  ruleStatement.AttrStrings("deps"),
				NixOpts:      ruleStatement.AttrStrings("nixopts"),
				BuildFile:    ruleStatement.AttrString("build_file"),
				Repositories: make(map[string]string),
			}
			*packages = append(*packages, pkg)
		}
	}
}

func initUpdateReposConfig(extensionConfig *config.Config, cexts []config.Configurer) {
	flagSet := flag.NewFlagSet("updateReposFlagSet", flag.ContinueOnError)

	for _, cext := range cexts {
		cext.RegisterFlags(flagSet, "update", extensionConfig)
	}

	for _, cext := range cexts {
		if err := cext.CheckFlags(flagSet, extensionConfig); err != nil {
			log.Fatal(err)
		}
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
