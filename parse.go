package main

import (
    "bufio"
    "fmt"
    "strconv"
)

type CharComp interface {
    includes(byte) bool
}

//A single character
type Char byte
func (c Char) includes(comp byte) bool { return byte(c) == comp }

//The starting and ending chars (inclusive) in a range of chars
type CharRange struct {
    start byte
    end   byte
}
func (cr CharRange) includes (comp byte) bool {
    return byte(cr.start) <= comp && byte(cr.end) >= comp
}

//A set of char comparables. Could contain other CharSets
type CharSet []CharComp
func (cs CharSet) includes (comp byte) bool {
    for _,c := range cs {
        if c.includes(comp) { return true }
    }
    return false
}

var ASCII = CharRange{0,127}

var Whitespace = CharSet{
                    CharRange{9,13}, //tab -> carriage return
                    Char(' '),
                 }

var Letters = CharSet{
                    CharRange{65,90},  //upper
                    CharRange{97,122}, //lower
              }

var Numbers = CharRange{48,57}

//For the simple string syntax (no d-quotes)
var SimpleChars = CharSet{
                    Letters,
                    Numbers,
                    Char('-'),
                    Char('_'),
                    Char('!'),
                    Char('?'),
                  }

//For translating the special characters
var SpecialCharsMap = map[byte]byte{
                        '\\': '\\',
                        '"' : '"',
                        '\'': '\'',
                        'n' : '\n',
                        'r' : '\r',
                        't' : '\t',
                        's' : ' ',
                        '0' : 0,
                      }

//For knowing what the other end of the seq should be
var SeqMap = map[byte]byte{
                '(': ')',
                '[': ']',
                '{': '}',
             }

//For simple unsigned integers
var UintChars = CharSet{
                     Numbers,
                     Char(','),
                }

//Looks at (but doesn't consume) the first byte in the buffer
func FirstByte(rbuf *bufio.Reader) (byte,error) {
    b,err := rbuf.Peek(1)
    if err != nil { return 0,err }
    return b[0],err
}

func PrettyChar(c byte) string {
    switch c {
        case '\n': return "newline"
        case '\t': return "tab"
        case '\r': return "carriage return"
        case 0: return "null byte"
        default: return string(c)
    }
}

func ExpectedPanic(rbuf *bufio.Reader,givenErr error,expected string) {
    if givenErr != nil {
        panic("Expected "+expected+", but found:"+givenErr.Error())
    }

    b,err := FirstByte(rbuf)
    if err != nil {
        panic("Expected "+expected+", but found:"+err.Error())
    } else {
        panic("Expected "+expected+", but found:"+PrettyChar(b))
    }
}

func ConcatByteSlices(a,b []byte) []byte {
    c := make([]byte,len(a)+len(b))
    copy(c,a)
    copy(c[len(a):],b)
    return c
}

func PullWhitespace(rbuf *bufio.Reader) (int,error) {
    var i int
    for i=0;;i++{
        b,err := FirstByte(rbuf)
        if err != nil { return i,err }
        if Whitespace.includes(b) != true { return i,nil }
        rbuf.ReadByte()
    }
    return i,nil
}

func PullUint(rbuf *bufio.Reader) ([]byte,error) {
    r := make([]byte,0,16)
    for {
        b,err := FirstByte(rbuf)
        if err != nil { return r,err }
        if UintChars.includes(b) != true { return r,nil }
        if b != ',' { r = append(r,b) }
        rbuf.ReadByte()
    }
}

func PullInteger(rbuf *bufio.Reader) ([]byte,error) {
    neg := false
    nb,err := FirstByte(rbuf)
    if err != nil { return nil,err }
    if nb == '-' {
        neg = true
        rbuf.ReadByte()
    }

    ui,err := PullUint(rbuf)
    if len(ui) == 0 || err != nil { return ui,err }

    var r []byte
    if neg {
        r = ConcatByteSlices([]byte{'-'},ui)
    } else {
        r = ui
    }

    return r,nil
    
}

func PullFullString(rbuf *bufio.Reader) ([]byte,error) {
    r := make([]byte,0,256)

    //Make sure string starts with dquote
    dq,err := FirstByte(rbuf)
    if err != nil { return nil,err }
    if dq != '"' { return r,nil }
    rbuf.ReadByte()
     
    for {
        b,err := FirstByte(rbuf)
        if err != nil { return r,err }
        if b == '"' {
            rbuf.ReadByte()
            return r,nil
        } else if b == '\\' {
            rbuf.ReadByte()
            ec,err := FirstByte(rbuf)
            if err != nil { return r,err }
            rbuf.ReadByte()

            if c,ok := SpecialCharsMap[ec]; ok {
                r = append(r,c)
            } else {
                r = append(r,ec)
            }
        } else {
            rbuf.ReadByte()
            r = append(r,b)
        }
    }
}

func PullSimpleString(rbuf *bufio.Reader) ([]byte,error) {
    r := make([]byte,0,16)
    
    for {
        b,err := FirstByte(rbuf)
        if err != nil { return r,err }
        if SimpleChars.includes(b) {
            rbuf.ReadByte()
            r = append(r,b)
        } else {
            return r,nil
        }
    }
}

func PullByte(rbuf *bufio.Reader) ([]byte,error) {
    sq,err := FirstByte(rbuf)
    if err != nil { return nil,err }
    if sq != '\'' { return []byte{},nil }
    rbuf.ReadByte()

    var rb byte
    b,err := FirstByte(rbuf)
    if err != nil { return nil,err }
    if b == '\\' {
        rbuf.ReadByte()
        ec,err := FirstByte(rbuf)
        if err != nil { return nil,err }

        if c,ok := SpecialCharsMap[ec]; ok {
            rb = c
        } else {
            return []byte{},nil
        }
    } else if ASCII.includes(b) {
        rb = b
    } else {
        return []byte{},nil
    }
    rbuf.ReadByte()
    rstr := strconv.Itoa(int(rb))
    return []byte(rstr),nil
}

func PullSeq(rbuf *bufio.Reader) ([]GngType,error) {
    d,err := FirstByte(rbuf)
    if err != nil { return nil,err }
    od,ok := SeqMap[d]
    if !ok { return nil,fmt.Errorf("Unknown seq type") }
    rbuf.ReadByte()

    r := make([]GngType,0,16)
    for {
        b,err := FirstByte(rbuf)
        if err != nil {
            return r,err
        } else if (b == od) {
            rbuf.ReadByte()
            return r,nil
        } else {
            el,err := PullElement(rbuf)
            if err != nil { return r,err }
            r = append(r,el)
        }
    }

}

func PullElement(rbuf *bufio.Reader) (GngType,error) {
    for {
        b,err := FirstByte(rbuf)
        if err != nil {
            return nil,err
        } else if Whitespace.includes(b) {
            rbuf.ReadByte() //ignore
        } else if Numbers.includes(b) || b == '-' {
            n,err := PullInteger(rbuf)
            if len(n) == 0 || err != nil { ExpectedPanic(rbuf,err,"number") }

            fb,err := FirstByte(rbuf)
            if err != nil { break }
            if fb == '.' {
                rbuf.ReadByte()
                ui,err := PullUint(rbuf)
                if len(ui) == 0 || err != nil { ExpectedPanic(rbuf,err,"number") }
                ui = ConcatByteSlices([]byte{'.'},ui)
                n = ConcatByteSlices(n,ui)
                return NewGngFloat(n),nil
            } else if fb == 'b' {
                rbuf.ReadByte()
                return NewGngByte(n),nil
            } else {
                return NewGngInteger(n),nil
            }

        } else if b == '"' {
            s,err := PullFullString(rbuf)
            if len(s) == 0 || err != nil { ExpectedPanic(rbuf,err,"end of string") }
            return NewGngString(s),nil
        } else if b == '\'' {
            b,err := PullByte(rbuf)
            if len(b) == 0 || err != nil { ExpectedPanic(rbuf,err,"character") }
            return NewGngByte(b),nil
        } else if Letters.includes(b) {
            s,err := PullSimpleString(rbuf)
            if len(s) == 0 || err != nil { ExpectedPanic(rbuf,err,"end of string") }
            return NewGngString(s),nil
        } else if _,ok := SeqMap[b]; ok {
            s,err := PullSeq(rbuf)
            if err != nil { ExpectedPanic(rbuf,err,"end of seq") }
            switch(b) {
                case '(': return NewGngList(s)
                case '[': return NewGngVector(s)
                case '{': return NewGngMap(s)
                default: panic("Unknown sequence type:"+string(b))
            }
        } else {
            ExpectedPanic(rbuf,nil,"start of valid data structure")
        }
    }
    return nil,nil
}
