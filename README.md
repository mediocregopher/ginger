# Ginger

Fibonacci function in ginger:

```
fib = {

    decr = { out = add < (in; -1;); };

    out = {

        n = tupEl < (in; 0;);
        a = tupEl < (in; 1;);
        b = tupEl < (in; 2;);

        out = if < (
            zero? < n;
            a;
            recur < (
                decr < n;
                b;
                add < (a;b;);
            );
        );

    } < (in; 0; 1;);
};
```

Usage of the function to generate the 6th fibonnaci number:

```
fib < 5;
```

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
