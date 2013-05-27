package main

import (
    "os"
    "bufio"
    "fmt"
)

func main() {
    file, err := os.Open("test.gng")
    if err != nil { panic(err) }
    rbuf := bufio.NewReader(file)

    for {
        el,err := PullElement(rbuf)
        if err != nil { break }
        fmt.Printf("%s\n",el)
    }
}
