# Notes

## http://nickdesaulniers.github.io/blog/2016/08/13/object-files-and-symbols/ (and it's follow up post)

These files are being used as an example:

```c
// main.c
// declare that these exist, but it's defined in hello.c
void hello();
void hello2();
int main() {
    hello();
    hello2();
    return 0;
}
```

```c
// hello.c
#include <stdio.h>
void hello() {
    puts("Hello World!");
}
```

```c
// hello2.c
#include <stdio.h>
void hello2() {
    puts("Hello Again!");
}
```

* Object files (.o)
    - `clang -c main.c hello.c hello2.c`
    - Contain the actual compiled machine code, but the addresses used need to
      be relocated when compiling the full binary.
    - Contain symbol table, which relates addresses to variables and functions
      defined in the object file.
    - `nm` can be used to inspect symbol table. Includes "undefined" symbols,
      which are symbols used by the object file but which aren't defined within
      it (and which are presumably defined elsewhere).
      ```
         ▻ nm main.o
                         U hello
                         U hello2
        0000000000000000 T main

         ▻ nm hello.o
        0000000000000000 T hello
        0000000000000000 r .L.str
                         U puts

         ▻ nm hello2.o
        0000000000000000 T hello2
        0000000000000000 r .L.str
                         U puts
      ```
    - `readelf` can be also used to dump the contents of the object file's
      symbol table on linux (`-s` displays symbol table):
      ```
         ▻ readelf -s main.o
        Symbol table '.symtab' contains 6 entries:
           Num:    Value          Size Type    Bind   Vis      Ndx Name
             0: 0000000000000000     0 NOTYPE  LOCAL  DEFAULT  UND
             1: 0000000000000000     0 FILE    LOCAL  DEFAULT  ABS main.c
             2: 0000000000000000     0 SECTION LOCAL  DEFAULT    2
             3: 0000000000000000     0 NOTYPE  GLOBAL DEFAULT  UND hello
             4: 0000000000000000     0 NOTYPE  GLOBAL DEFAULT  UND hello2
             5: 0000000000000000    37 FUNC    GLOBAL DEFAULT    2 main

         ▻ readelf -s hello.o
        Symbol table '.symtab' contains 6 entries:
           Num:    Value          Size Type    Bind   Vis      Ndx Name
             0: 0000000000000000     0 NOTYPE  LOCAL  DEFAULT  UND
             1: 0000000000000000     0 FILE    LOCAL  DEFAULT  ABS hello.c
             2: 0000000000000000    13 OBJECT  LOCAL  DEFAULT    4 .L.str
             3: 0000000000000000     0 SECTION LOCAL  DEFAULT    2
             4: 0000000000000000    29 FUNC    GLOBAL DEFAULT    2 hello
             5: 0000000000000000     0 NOTYPE  GLOBAL DEFAULT  UND puts
      ```

* Static library files (.a)
    - `ar` utility creates uncompressed, static (those might be synonomous in
      this context?) archives, with the `.a` extension.
    - In the context of compiling code, `.a` files are archives of multiple
      object files, with the symbol table preserved in a way where nm and ilk
      can still understand it.
      ```
         ▻ ar r hello.a hello.o hello2.o
         ▻ nm hello.a

        hello.o:
        0000000000000000 T hello
        0000000000000000 r .L.str
                         U puts

        hello2.o:
        0000000000000000 T hello2
        0000000000000000 r .L.str
                         U puts
      ```
    - This `.a` file can then be passed into clang as if it was an object file,
      and the resulting binary would statically contain all symbols from the
      archive that it needs:
      ```
         ▻ clang main.o hello.a
         ▻ ./a.out
        Hello World!
        Hello Again!
      ```

* Dynamic/shared library files (.so on Linux, .dylib on OSX, .dll on Windows)
    - If multiple programs share the same library and are being statically
      compiled then, when run, that library ends up in memory twice. Dynamic
      linking allows the library to be dynamically linked in at runtime, to save
      memory use.
    - Can be compiled from either source or object files:
        - `clang -shared hello.c hello2.c -o hello.so`
        - `clang -shared hello.o hello2.o -o hello.so`
    - Then used in final compilation normally: `clang main.o ./hello.so`
      ```
         ▻ clang main.o ./hello.so
         ▻ ldd a.out
                linux-vdso.so.1 (0x00007ffd5665c000)
                ./hello.so (0x00007f7f9bbe8000)
                libc.so.6 => /usr/lib/libc.so.6 (0x00007f7f9b830000)
                /lib64/ld-linux-x86-64.so.2 => /usr/lib64/ld-linux-x86-64.so.2 (0x00007f7f9bfec000)
      ```
    - `strace ./a.out` can be used to view all system calls a binary makes while
      it runs, including opening and reading dynamic libaries, which will look
      like:
      ```
        openat(AT_FDCWD, "./hello.so", O_RDONLY|O_CLOEXEC) = 3
        read(3, "\177ELF\2\1\1\0\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0\360\4\0\0\0\0\0\0"..., 832) = 832
        fstat(3, {st_mode=S_IFREG|0755, st_size=7784, ...}) = 0
        mmap(NULL, 8192, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x7f58fc93a000
      ```
    - `LD_DEBUG` env variable can also be used for tracing information, as well
      as dynamic library search path stuff.
    - If hello.so were placed into `/usr/lib` or `/usr/local/lib` then
      compilation could be done with just `clang hello.o -lhello`. `-L` can add
      library search paths as well.
    - `pkg-config` and its associated `.pc` files can be used by library authors
      to specify flags required when compiling using that shared library (e.g.
      to include necessary header files and whatnot).
    - `LD_PRELOAD` can be used to pre-eminently link in a shared library which
      will get searched first before symbols from the "real" shared libraries
      are searched, allowing for code-replace and such.

## http://www.linuxjournal.com/article/1059

* ELF Header
    - Contains (I think) information on sections, their sizes, and their offsets

* ELF File Sections
    - After the header, ELF files composed of multiple sections
    - Each section is composed of information of a similar type
    - All sections are loaded into memory, presumably, but certain are loaded
      into read-only pages (like .text, the executable code) and others are
      loaded into read-write.
    - It would seem that the read-only/read-write dichotemy is enforced by the
      "memory manager", which is part of the cpu?
    - Different sections:
        - .text (ro): executable code
        - .data (rw): variables the user has specified an initial value for
        - .bss (rw): variables the user has not specified an initial value for,
          separate from .data because there's no need to waste space in the
          binary file with zeros.
        - symbol table for debugging (and possibly dynamic linking?)

* Shared libraries
    - so's are designed to be "position independent", meaning when the so is
      loaded at binary runtime the place in memory that it is loaded into is not
      actually important. The `-fPIC` compiler option used-to-be/is (?)
      important in order to enable this. (PIC being "position independent code")
    - Compiler reserves a register which points to the start of a "global offset
      table", which is used to support global variables within shared libraries
      using PIC. (I guess shared library global vars are shared across
      processes?)
    - Procedure Linkage Table is like the GOT but for functions, it's basically
      a jump table within the library file. If the user wants to redefine one of
      the shared library's functions, and have all other functions within the so
      use that new one, then the PLT entry for that function is the only need
      which gets changed.

* Compiling
    - During compilation the compiler will keep track of symbols needing
      "relocating", meaning they are external to the object file and will need
      to be patched in during linking. Each relocated symbol is marked as such
      in the symbol table (I think), along with the offset into .text where that
      symbol was used, and where the linker needs to place the actual address.

## https://blog.oracle.com/ksplice/hello-from-a-libc-free-world-part-2

With the following file:
```c
// main.c
void alloc_boi() {
    char *str = "Hello world";
}

void _start() {
    alloc_boi();
    asm("movl $1, %eax;" // what
        "movl $0, %ebx;"
        "int $0x80;");
}
```

* Compile the above with `clang -nostdlib main.c`, the result will have very
  little in it, but it seems there's still some extra uneeded sections which
  could be removed.

* `main.c` uses `_start` instead of `main` since that's the actual first
  function called, but normally it gets filled with libc junk (like importing
  environment variables and such).

* The exit call is needed to be explicitly defined otherwise the process won't
  exit, instead execution will run past `.text` and segfault.

* The disassembly from above looks like this:
  ```
  Disassembly of section .text:

  0000000000000250 <alloc_boi>:
   250:   48 8d 05 1d 00 00 00    lea    0x1d(%rip),%rax        # 274 <_start+0x14>
   257:   48 89 44 24 f8          mov    %rax,-0x8(%rsp)
   25c:   c3                      retq
   25d:   0f 1f 00                nopl   (%rax)

  0000000000000260 <_start>:
   260:   50                      push   %rax
   261:   e8 ea ff ff ff          callq  250 <alloc_boi>
   266:   b8 01 00 00 00          mov    $0x1,%eax
   26b:   bb 00 00 00 00          mov    $0x0,%ebx
   270:   cd 80                   int    $0x80
   272:   58                      pop    %rax
   273:   c3                      retq
  ```
   The `alloc_boi` section is the interesting one:

  - `%rsp` is the stack pointer, `%rbp` is apparently general purpose but is
    used in this context as the "frame pointer", meaning the start of the stack
    frame.  This is a small optimization which allows referencing memory from a
    point which is constant during the function (the frame's start) rather than
    a point which changes (the stack pointer's position). This optimization can
    be negated by compiling with `-fomit-frame-pointer`

  - It also contains `lea    0x1d(%rip),%rax` at the top of `alloc_boi`'s
    disassembly. `lea` is "load effective address". Basically puts the
    calculated pointer into `%rax`. The pointer being calculated is
    `0x1d(%rip)`, which is the instruction pointer + 0x12. The instruction
    pointer's value is always the next instruction to be run, and in this case
    is `0x257`. Adding `0x1d` to that gives `0x274`, which is the first byte in
    the `.rodata` section, the start of the `Hello World` string.

  - The subsequent `mov    %rax,-0x8(%rsp)` is moving the pointer (stored in
    `%rax`) and putting it onto the stack.

## VLA

With the following file:

```c
// main.c
void do_the_thing(int n) {
    int arr[n];
    for (int i=0; i<n; i++) {
        arr[i] = i;
    }
    asm("nop; nop; nop;");
}


void _start() {
    do_the_thing(10);
    asm("movl $1, %eax;"
        "movl $0, %ebx;"
        "int $0x80;");
}
```

And compiled into LLVM IR with the following:

```
clang -nostdlib -fno-stack-protector -fomit-frame-pointer -S -emit-llvm main.c
```

We can see how llvm handles VLA. It ain't pretty, that's for sure.
