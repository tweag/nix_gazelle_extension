`vanilla` workspace represents the most popular structure of nix+Bazel codebase, where every derivation is a separate entity.

---
Let's take a look at the setup:
```
cat WORKSPACE
```
---
Top level `BUILD` file contains a convenience macro wrapping `gazelle`, and extension specific directives.
* `gazelle:nix_repositories` defines mapping between
     - nix search path
     - `nixpkgs_local_repository` target name
     - path to a nix expression representing the content of Nixpkgs
```
cat BUILD.bazel
```
---
`folks` directory contains three nix expressions:
```
tree folks
```
---
`i-need-a-friend` expression simulates build process of a binary:
```
cat folks/i-need-a-friend/default.nix
```
---
`cowsay` expression is - surprisingly - building cowsay binary.
```
cat folks/cowsay/default.nix
```
---
To generate Bazel targets, invoke `gazelle` like so:
```
bazel run //:gazelle-update-all
```
---
To see more detailed output set `GAZELLE_LANGUAGES_NIX_LOG_LEVEL` environment variable:
```
GAZELLE_LANGUAGES_NIX_LOG_LEVEL=debug bazel run //:gazelle-update-all
```
---
Once generation finishes, you should see three new `BUILD.bazel` files:
```
git status folks
```
---
Generated `BUILD.bazel` files contain two targets:
- `filegroup`: exposes sources to which we will refer from a `WORKSPACE`
- `nixpkgs_package_manifest`: a placeholder, unused in build context
```
cat folks/i-need-a-friend/BUILD.bazel
```
---
Generated `BUILD.bazel` files contain two targets:
- `filegroup`: exposes sources to which we will refer from a `WORKSPACE`
- `nixpkgs_package_manifest`: a placeholder, unused in build context
```
cat folks/cowsay/BUILD.bazel
```
---
Repositories are generated in a `WORKSPACE` file
```
diff WORKSPACE
```
---
Generated target can be verified with:
```
bazel run @folks.cowsay//:bin/cowsay -- 'Hello, Bazel'
```
