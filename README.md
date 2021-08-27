# Ginger

Fibonacci function in ginger:

```
fib {
    decr { out add(in, -1) }

    out {
        n 0(in),
        a 1(in),
        b 2(in),

        out if(
            zero?(n),
            a,
            recur(decr(n), b, add(a,b))
        )

    }(in, 0, 1)
}
```

Usage of the function to generate the 6th fibonnaci number:

```
fib(5)
```
