package gazelle

const (
	LANGUAGE_NAME = "nix"
	EXPORT_RULE   = "filegroup"
	MANIFEST_RULE = "nixpkgs_package_manifest"
	PACKAGE_RULE  = "nixpkgs_package"

	// Nix2BuildPath path to a nix evaluator binary.
	FPTRACE_PATH = "external/fptrace/bin/fptrace"
)
