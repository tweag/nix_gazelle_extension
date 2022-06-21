`readtree` workspace represents a more sophisticated way of integrating nix+Bazel codebase. It takes a leverage of an additional entrypoint expression reffered to as "prelude".

Compared to `vanilla` example, `readtree` differs in the project structure and the way the derivations are built. 

```
diff {../vanilla/,./}folks/i-need-a-friend/default.nix
cat default.nix
```

---
Let's generate Bazel definitions for the code base
```
bazel run //:gazelle-update-all
```
---
`BUILD` files created:
```
git status folks
```
---
Definitions of generated `nixpks_package(s)` in `WORKSPACE`:
```
git diff WORKSPACE
```
---
Asking Bazel to build a particular `nixpkgs_package`
```
bazel build @folks.cowsay//:bin/cowsay
```
Building the very same derivation with plain `nix-build`
```
nix-build -A folks.cowsay
```
---
The `@folks.cowsay` repository contents
```
tree -l bazel-readtree/external/folks.cowsay
```
---
The `folks.cowsay` deriviation output path contents
```
tree -l result
```
---
Building a deriviation directly from `nixpkgs` is also possible
```
nix-build -A neofetch
```
