def export_nix(name = "export_nix", files = [], deps = [], **kwargs):
    """ This macro makes it possible for gazelle to store additional information. """
    native.exports_files(files)
