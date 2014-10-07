package parse

import (
	"bufio"
	"io"
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
func ReadString(r io.Reader) (types.Str, error) {
	buf := bufio.NewReader(r)
	str := types.Str("\"")
	for {
		piece, err := buf.ReadBytes('"')
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
