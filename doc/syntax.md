# Syntax

Ginger is a lisp language, so knowing the syntax is as simple as knowing the data structures.

## Strings

Strings are declared two different ways. The standard way, with double quotes:
```
"this is a string\n <- that was a newline, \t \r \0 \" \\ also work
literal whitespace characters are properly parsed as well
\s is a space"
```

The second way only works if your string contains exclusively the following characters:
`a-z A-Z 0-9 _ - ! ? . /` (spaces added for readability)
```
neat
this_works
so-does-this!
what-about-this?_YUP!
/directory/with/a/file.ext
```

## Integers

Integers are defined the standard way, a bunch of numbers with an optional negative. The only
interesting thing is that commas inside the number are ignored, so you can make your literals pretty:
```
0
1
-2
4,000,000
```

## Floats

Pretty much the same as integers, but with a period thrown in there. If there isn't a period, it's not
a float:
```
0.0
-1.5
-1,003.004,333,203
```

## Bytes

Singular unsigned bytes, are also supported. There are two ways to declare them. With a trailing `b`:
```
0b
10b
255b
```

and with a `'` followed by a character (or escaped character):
```
'c
'h
'\n
'\\
''
```

## Vectors

A vector is a sequence of elements wrapped in `[ ... ]` (no commas):
```
[ a b 0 1 2
    [embedded also works]
]
```

## Lists

A list is a sequence of elements wrapped in `( ... )` (no commas):
```
( a b 0 1 2
    [embedded also works]
    (and mixed types)
)
```

## Maps

A map is a sequence of elements wrapped in `{ ... }`. There must be an even number of elements, and
there is no delimeter between the keys and values. Keys can be any non-sequence variable, values can
be anything at all:
```
{ a 1
  b 2
  c [1 2 3]
  d (four five six)
  e { 7 seven } }
```

## Comments

A semicolon delimits the beginning of a comment. Anything after the semicolon till the end of the line
is discarded by the parser:
```
;I'm a comment
"I'm a string" ;I'm another comment!
```
