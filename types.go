package main

import (
    "strconv"
    "fmt"
)

type GngType interface {}

type GngString string
func NewGngString(b []byte) GngString { return GngString(b) }

type GngByte byte
func NewGngByte(b []byte) GngByte {
    i,err := strconv.Atoi(string(b))
    if err != nil { panic(err) }
    return GngByte(i)
}

type GngInteger int64
func NewGngInteger(b []byte) GngInteger {
    i,err := strconv.Atoi(string(b))
    if err != nil { panic(err) }
    return GngInteger(i)
}

type GngFloat float64
func NewGngFloat(b []byte) GngFloat {
    f,err := strconv.ParseFloat(string(b),64)
    if err != nil { panic(err) }
    return GngFloat(f)
}

type GngVector []GngType
func NewGngVector(e []GngType) (GngVector,error) { return GngVector(e),nil }

type GngList []GngType
func NewGngList(e []GngType) (GngList,error) { return GngList(e),nil }

type GngMap []GngType
func NewGngMap(e []GngType) (GngMap,error) {
    if len(e)%2 != 0 {
        return nil,fmt.Errorf("uneven number of elements in map literal")
    } else {
        return GngMap(e),nil
    }
}
