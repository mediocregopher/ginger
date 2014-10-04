# Hello world

The following code in a file called `hello-world.gg` can be compiled to a static
binary which will output "Hello World!" and exit:

```
(. package "github.com/mediocregopher/ginger-hello-world"
    (. defn main []
        (: fmt.Println "Hello World!"))
)
```

While in the same directory as `hello-world.gg`, this can be compiled and run
with:

```
ginger run
```

and built into a static binary with:

```
ginger build
```

## package

The package string defines where in the namespace tree this file belongs. Code
in this file can directly reference variables, private or public, from other
files in the same package.

The physical layout of the files is not consequential, ginger only cares what
package they say they belong to, it's not required that the directory tree
matches the namespace tree.

## main

Every executable which ginger compiles needs one `main` function to use as the
entrypoint. Ginger looks for a `main` function in all the `.gg` files in the
current working directory in order to compile. If it finds one it uses that, if
it finds zero or more than one it errors and tells the user they need to specify
either a package or file name.

## Compilation

Ginger is first compiled into go code, and the subsequently uses the go tool to
compile that code into a static binary. Because of this ginger is able to use
any code from the go standard library, as well as any third-party go libraries
which may want to be used.

Ginger takes advantage of the `GOPATH` when compiling in order to find packages.
Upon starting up ginger will set the `GOPATH` as follows for the duration of the
run:

```
GOPATH=./target:$GOPATH
```

The following steps are then taken for compilation:

* Create a folder in the cwd called `target`, if it isn't already there. Also
  create a `src` subfolder inside of that.

* Scan the cwd and any subdirectories (including `target`) for `.gg` files.
  Translate them into `.go` files and place them according to their package name
  in the `target/src` directory. So if a file has a package
  `github.com/mediocregopher/foobar` it would be compiled and placed as
  `target/src/github.com/mediocregopher/foobar/file.go`. The `main` function is
  found and the package for it is determined in this step as well.

* As the cwd `.gg` files are scanned a set of packages which are required is
  built. If any are not present in `target` at the end of the last step and are
  not present as go projects in the `GOPATH` then they are searched for as
  ginger projects in the `GOPATH`. Packages found will be translated into the
  `target` directory (very important, this prevents the global GOPATH from
  getting cluttered with tranlated `.go` files which aren't actually part of the
  project). This step is repeated until all dependencies are met, or until
  they're not and an error is thrown.

* At this point all necessary `.go` files to build the project are present in
  the `GOPATH` (global or `target`). `go build` is called on the `main` package
  and output to the `target` directory.

### Properties

Given those compilation steps ginger has the following properties:

* Dependencies for a ginger project, either go or ginger, can be installed
  globally or to the `target` folder (sandboxed) by a dependency management tool
  (probably built into ginger).

* It is easy to find exactly what your code is being translated to since it will
  always be in the `target` directory.

* Translated code will not clutter up the global GOPATH.

* Only the `target` directory needs to be added to a `.gitignore` file.

* Some amount of ginger-to-ginger monkey patching may be possible. Not sure if
  this is a good or bad thing.

* Compilation may take a while. There is some amount of hunting for `.gg` files
  required, and all found ones *must* be compiled even if they're not going to
  be used, since package names can be arbitrarily stated. There are some ways to
  help this:

    * Might be worth taking a shortcut like grepping through all files in the
      `GOPATH` for the package string only compiling the ones which pass the
      grep.

    * Put placeholder `.go` files in the `target` directory to indicate that the
      package isn't needed for subsequent installs. Not the *best* idea, since
      changes to the dependency list in the project may not correctly process.
