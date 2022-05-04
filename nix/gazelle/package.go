package gazelle

type nixPackage struct {
	name, nixFile, buildFile    string
	files, nixFileDeps, nixOpts []string
	repositories                map[string]string
}
