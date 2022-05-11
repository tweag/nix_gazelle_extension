package gazelle

const (
	languageName = "nix"
	exportRule   = "export_nix"
	manifestRule = "nixpkgs_package_manifest"
	packageRule  = "nixpkgs_package"

	// Nix2BuildPath path to a nix evaluator binary.
	nix2BuildPath = "external/fptrace/bin/fptrace"
)
