# Ginger

A programming language utilizing a graph datastructure for syntax. Currently in
super-early-alpha-don't-actually-use-this-for-anything development.

## Development

Current efforts on ginger are focused on a golang-based virtual machine, which
will then be used to bootstrap the language. go >=1.18 is required for this vm.

If you are on a linux-amd64 machine with nix installed, you can run:

```
nix-shell -A shell
```

from the repo root and you will be dropped into a shell with all dependencies
(including the correct go version) in your PATH, ready to use. This could
probably be expanded to other OSs/architectures easily, if you care to do so
please check out the `default.nix` file and submit a PR!

## Demo

An example program which computes the Nth fibonacci number can be found at
`examples/fib.gg`. You can try it out by doing:

```
go run ./cmd/eval/main.go "$(cat examples/fib.gg)" 5
```

Where you can replace `5` with any number. The vm has only been given enough
capability to run this program as a demo, and is extremely poorly optimized (as
will be evident if you input any large number). Further work is obviously TODO.
