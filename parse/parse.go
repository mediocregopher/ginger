package parse

import (
	"bufio"
	"strconv"

	"github.com/mediocregopher/ginger/types"
)

//func ReadElem(r io.Reader) (types.Elem, error) {
//	buf := bufio.NewReader(r)
//	var err error
//	for {
//	}
//}

// ReadString reads in a string from the given reader. It assumes the first
// double-quote has already been read off. Ginger strings are wrapped with " and
// are allowed to have newlines literal in them. In all other respects they are
// the same as go strings.
func ReadString(r *bufio.Reader) (types.Str, error) {
	str := types.Str("\"")
	for {
		piece, err := r.ReadBytes('"')
		if err != nil {
			return "", err
		}
		str += types.Str(piece)
		if piece[len(piece)-2] != '\\' {
			break
		}
	}

	ret, err := strconv.Unquote(string(str))
	if err != nil {
		return "", err
	}
	return types.Str(ret), nil
}


// Returns (isNumber, isFloat). Can never return (false, true)
func whatNumber(el string) (bool, bool) {
	var isFloat bool
	first := el[0]

	var start int
	if first == '-' {
		if len(el) == 1 {
			return false, false
		}
		start = 1
	}

	el = el[start:]
	for i := range el {
		if el[i] == '.' {
			isFloat = true
		} else if el[i] < '0' || el[i] > '9' {
			return false, false
		}
	}

	return true, isFloat
}

// Given a string with no spaces and with a length >= 1, parses it into either a
// number or string.
func ParseBareElement(el string) (types.Elem, error) {
	isNumber, isFloat := whatNumber(el)
	if isNumber {
		if isFloat {
			f, err := strconv.ParseFloat(el, 64)
			if err != nil {
				return nil, err
			}
			return types.Float(f), nil
		} else {
			i, err := strconv.ParseInt(el, 10, 64)
			if err != nil {
				return nil, err
			}
			return types.Int(i), nil
		}
	}

	if el[0] == ':' {
		return types.Str(el), nil
	}

	return types.Str(":"+el), nil
}
