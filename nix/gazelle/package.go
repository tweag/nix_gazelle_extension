package gazelle

type nixPackage struct {
	name, nixFile, buildFile, rel string
	files, nixFileDeps, nixOpts   []string
	repositories                  map[string]string
}
