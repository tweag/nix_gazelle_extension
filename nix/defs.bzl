def export_nix(name = "export_nix", files = []):
    """ This macro makes it possible for gazelle to store additional information. """
    native.exports_files(files)

def nixpkgs_package_manifest(**kwargs):
    pass
