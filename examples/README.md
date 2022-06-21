## Examples

The `examples` directory contains two Bazel Workspaces - each of which represents a different approach to structuring nix codebase and integrating it with Bazel.   
By default, the directory is used in gazelle-extenion testing suite, so we will need to do a bit of magic first.

Let's convert the testing suite into standard Bazel workspace:

```
bazel run //examples:generate
```

**Notice**

Presence of any `BUILD{.bazel}` file in the root of test workspace(s), will cause the gazelle tests not to find this example.  
To run the code manually, run `//examples:generate` first, then enter one of test workspaces and issue `//:gazelle-update-all`

## Testing
As mentioned at in the previous section, `examples` package also specifies a gazelle-extension testing suite.  
One can verify if the execution of an extension results in appropriate definitions being created.

```
git clean -xdfq examples
git checkout -- examples

bazel test //examples/...
```
