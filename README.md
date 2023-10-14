# Ginger

A programming language utilizing a graph datastructure for syntax. Currently in
super-early-alpha-don't-actually-use-this-for-anything development.

## Development

Current efforts on ginger are focused on a golang-based virtual machine, which
will then be used to bootstrap the language.

If you are on a machine with nix installed, you can run:

```
nix develop
```

from the repo root and you will be dropped into a shell with all dependencies
(including the correct go version) in your PATH, ready to use.

## Demo

An example program which computes the Nth fibonacci number can be found at
`examples/fib.gg`. You can try it out by doing:

```
go run ./cmd/eval/main.go "$(cat examples/fib.gg)" 5
```

Where you can replace `5` with any number.
