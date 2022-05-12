package gazelle

const (
	LANGUAGE_NAME = "nix"
	EXPORT_RULE   = "export_nix"
	MANIFEST_RULE = "nixpkgs_package_manifest"
	PACKAGE_RULE  = "nixpkgs_package"

	// Nix2BuildPath path to a nix evaluator binary.
	NIX2BUILDPATH = "external/fptrace/bin/fptrace"
)
