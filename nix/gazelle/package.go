package gazelle

// FIXME: nixPackage should not carry information
// meant for build context target(s) i.e. nix_export
type nixPackage struct {
	name, nixFile, buildFile    string
	nixFileDeps, files, nixOpts []string
	repositories                map[string]string
}

type nixExport struct {
	name  string
	files []string
}
