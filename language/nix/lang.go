package gazelle_nix

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

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
)

const nixName = "gazelle_nix"

var _ = fmt.Printf

type nixLang struct{}

// NewLanguage implementation
func NewLanguage() language.Language {
	return &nixLang{}
}

func (l *nixLang) Name() string { return nixName }

// Kinds returns a map of maps rule names (kinds) and information on how to
// match and merge attributes that may be found in rules of those kinds. All
// kinds of rules generated for this language may be found here.
func (l *nixLang) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		"export": {
			MatchAny:   false,
			MatchAttrs: []string{"name"},
		},
		"nixpkgs_package": {
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
			Name: "@io_tweag_gazelle_nix//tools:exporter.bzl",
			Symbols: []string{
				"export",
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
	nixPreludeConf := args.Config.Exts[nixName].(Config).NixPrelude

	nixFiles := make(map[string]string)
	nixNames := make(map[string]string)
	nixFilesDeps := make(map[string]*DepSets)
	for _, f := range append(args.RegularFiles, args.GenFiles...) {
		if !strings.HasSuffix(f, "default.nix") {
			continue
		}

		pth := filepath.Join(args.Dir, f)

		log.Printf("parsing nix file: path=%q", pth)
		nixFileDep, err := nixToDepSets(nixPreludeConf, pth)
		if err != nil {
			log.Printf("failed parsing nix file: path=%q", pth)
			continue
		}

		nixFiles[f] = pth
		nixNames[f] = args.Rel
		nixFilesDeps[f] = nixFileDep
	}

	libraries := make(map[string]*nixLibrary)

	for nixFile := range nixFiles {
		fileDeps := nixFilesDeps[nixFile]
		lib := libraries[nixFile]
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
		lib = &nixLibrary{
			Name:      pkgName,
			NixFile:   tgtName,
			Files:     fileDeps.DepSets[1].Files,
			Deps:      fileDeps.DepSets[0].Files,
			NixOpts:   nixOpts,
			BuildFile: buildFile,
		}
		libraries[nixFile] = lib
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
	r := rule.NewRule("export", l.Name)
	r.SetAttr("nix_file", l.NixFile)
	sort.Strings(l.Files)
	r.SetAttr("files", l.Files)
	r.SetAttr("deps", l.Deps)
	r.SetAttr("nixopts", l.NixOpts)
	r.AddComment("# autogenerated")
	if len(l.BuildFile) > 0 {
		r.SetAttr("build_file", l.BuildFile)
	}
	return r
}

func fixGazelle(c *config.Config, f *rule.File) {
	for _, l := range f.Loads {
		if l.Has("export") {
			l.Remove("export")
			if l.IsEmpty() {
				l.Delete()
			}
		}
	}
	var nixRules []*rule.Rule
	for _, r := range f.Rules {
		if r.Kind() == "export" {
			nixRules = append(nixRules, r)
		}
	}

	for _, r := range nixRules {
		r.Delete()
	}
}

// Fix repairs deprecated usage of language-specific rules in f. This is
// called before the file is indexed. Unless c.ShouldFix is true, fixes
// that delete or rename rules should not be performed.
func (l *nixLang) Fix(c *config.Config, f *rule.File) {
	fixGazelle(c, f)
}

func (*nixLang) CheckFlags(
	fs *flag.FlagSet,
	c *config.Config,
) error {
	return nil
}

// Config configuration for language extension
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

func (*nixLang) Configure(c *config.Config, rel string, f *rule.File) {
	if f == nil {
		return
	}

	m, ok := c.Exts[nixName]
	var extraConfig Config
	if ok {
		extraConfig = m.(Config)
	} else {
		extraConfig = Config{
			NixPrelude:      "",
			NixRepositories: make(map[string]string),
		}
	}
	for _, directive := range f.Directives {
		switch directive.Key {
		case "nix_prelude":
			extraConfig.NixPrelude = directive.Value
		case "nix_repositories":
			parseNixRepositories(&extraConfig, directive.Value)
		}
	}
	c.Exts[nixName] = extraConfig
}

func parseNixRepositories(nixconfig *Config, value string) {
	r := strings.Split(value, " ")
	kv := make([]string, 0, len(r)*2)
	for _, key := range r {
		kv = append(kv, strings.Split(key, "=")...)
	}
	if len(kv)%2 != 0 {
		msg := "Can't parse value of nix_repositories: %s"
		err := fmt.Errorf(msg, value)
		log.Fatal(err)
	}
	for i := 0; i < len(kv); i += 2 {
		nixconfig.NixRepositories[kv[i]] = kv[i+1]
	}
}

// Embeds returns a list of labels of rules that the given rule
// embeds. If a rule is embedded by another importable rule of the
// same language, only the embedding rule will be indexed. The
// embedding rule will inherit the imports of the embedded rule.
func (l *nixLang) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

// checkPrefix checks that a string may be used as a prefix. We forbid local
// (relative) imports and those beginning with "/". We allow the empty string,
// but generated rules must not have an empty importpath.
func checkPrefix(prefix string) error {
	if strings.HasPrefix(prefix, "/") || build.IsLocalImport(prefix) {
		return fmt.Errorf("invalid prefix: %q", prefix)
	}
	return nil
}

// Imports returns a list of ImportSpecs that can be used to import
// the rule r. This is used to populate RuleIndex.
//
// If nil is returned, the rule will not be indexed. If any non-nil
// slice is returned, including an empty slice, the rule will be
// indexed.
func (l *nixLang) Imports(
	c *config.Config,
	r *rule.Rule,
	f *rule.File,
) []resolve.ImportSpec {
	var prefix string
	switch r.Kind() {
	case "export":
		prefix = "exports:"
	case "nixpkgs_package":
		prefix = "nixpkgs_package:"
	}
	return []resolve.ImportSpec{{nixName, prefix + r.Name()}}
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
	nixRepositoriesConf := args.Config.Exts[nixName].(Config).NixRepositories

	theRules := make([]*rule.Rule, len(packageList))
	for i, pkg := range packageList {
		r := rule.NewRule("nixpkgs_package", pkg.Name)
		r.SetAttr("nix_file", pkg.NixFile)
		r.SetAttr("nixopts", pkg.NixOpts)
		r.SetAttr("nix_file_deps", pkg.NixFileDeps)
		r.SetAttr("repositories", nixRepositoriesConf)
		if len(pkg.BuildFile) > 0 {
			r.SetAttr("build_file", pkg.BuildFile)
		}
		r.AddComment("# autogenerated")
		theRules[i] = r
	}

	return language.UpdateReposResult{
		Gen: theRules,
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

// DepSets collection of DepSet structs
type DepSets struct {
	DepSets []DepSet
}

// DepSet represents dependencies of this package
type DepSet struct {
	Kind  string
	Files []string
}

// Nix2BuildPath path to a nix evaluator binary
const Nix2BuildPath = "external/fptrace/bin/fptrace"

func reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func nixToDepSets(nixPrelude, nixFile string) (*DepSets, error) {
	wsroot := os.Getenv("BUILD_WORKSPACE_DIRECTORY")
	scanNix, err := bazel.Runfile(Nix2BuildPath)
	tmpfile, err := ioutil.TempFile("", "nix-gzl*.json")
	if err != nil {
		log.Fatal(err)
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
		log.Printf(
			"\033[31m" + string(
				out,
			) + "\n" + string(
				debug.Stack(),
			) + "\033[0m",
		)
		fmt.Errorf("%v", err)
		return nil, err
	}

	var traceOuts []TraceOut
	byteValue, _ := os.ReadFile(tmpfile.Name())
	err = json.Unmarshal(byteValue, &traceOuts)
	if err != nil {
		fmt.Errorf("%v", err)
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

	for _, x := range packages {
		temp := filteredFiles[:0]
		for _, y := range filteredFiles {
			if strings.HasPrefix(y, x) {
				pkg := strings.TrimSuffix(
					strings.TrimPrefix(x, wsroot+"/"),
					"/",
				)
				reltarget := strings.TrimPrefix(y, x)
				target := "//" + pkg + ":" + reltarget
				targets = append(targets, target)
			} else {
				temp = append(temp, y)
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
	c *config.Config,
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

	initUpdateReposConfig(c, cexts)

	walk.Walk(
		c,
		cexts,
		[]string{},
		walk.VisitAllUpdateDirsMode,
		func(
			dir,
			rel string,
			c *config.Config,
			update bool,
			f *rule.File,
			subdirs,
			regularFiles,
			genFiles []string,
		) {
			collectDependenciesFromFile(f, c, &packages)
		},
	)
	return packages
}

func collectDependenciesFromFile(
	f *rule.File,
	c *config.Config,
	packages *[]nixWorkspaceLibrary,
) {
	if f == nil {
		return
	}

	for _, r := range f.Rules {
		if r.Kind() == "export" {
			pkg := nixWorkspaceLibrary{
				Name:        r.AttrString("name"),
				NixFile:     r.AttrString("nix_file"),
				NixFileDeps: r.AttrStrings("deps"),
				NixOpts:     r.AttrStrings("nixopts"),
				BuildFile:   r.AttrString("build_file"),
			}
			*packages = append(*packages, pkg)

		}
	}
}

func initUpdateReposConfig(c *config.Config, cexts []config.Configurer) {
	fs := flag.NewFlagSet("updateReposFlagSet", flag.ContinueOnError)

	for _, cext := range cexts {
		cext.RegisterFlags(fs, "update", c)
	}

	for _, cext := range cexts {
		if err := cext.CheckFlags(fs, c); err != nil {
			log.Fatal(err)
		}
	}
}

type resolver struct{}

// Name returns the name of the language. This should be a prefix of the
// kinds of rules generated by the language, e.g., "go" for the Go extension
// since it generates "go_library" rules.
func (resolver) Name() string {
	return "go"
}

// Imports returns a list of ImportSpecs that can be used to import the rule
// r. This is used to populate RuleIndex.
//
// If nil is returned, the rule will not be indexed. If any non-nil slice is
// returned, including an empty slice, the rule will be indexed.
func (resolver) Imports(
	c *config.Config,
	r *rule.Rule,
	f *rule.File,
) []resolve.ImportSpec {
	fmt.Println("Imports:", r.Kind(), r.Name())
	return nil
}

// Embeds returns a list of labels of rules that the given rule embeds. If
// a rule is embedded by another importable rule of the same language, only
// the embedding rule will be indexed. The embedding rule will inherit
// the imports of the embedded rule.
func (resolver) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

// Resolve translates imported libraries for a given rule into Bazel
// dependencies. A list of imported libraries is typically stored in a
// private attribute of the rule when it's generated (this interface doesn't
// dictate how that is stored or represented). Resolve generates a "deps"
// attribute (or the appropriate language-specific equivalent) for each
// import according to language-specific rules and heuristics.
func (resolver) Resolve(
	c *config.Config,
	ix *resolve.RuleIndex,
	rc *repo.RemoteCache,
	r *rule.Rule,
	from label.Label,
) {
	fmt.Println("Resolve:", r.Name(), "from", from)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
