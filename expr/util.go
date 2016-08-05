package expr

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
)

func randStr() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func exprsJoin(ee []Expr) string {
	strs := make([]string, len(ee))
	for i := range ee {
		strs[i] = fmt.Sprint(ee[i].Actual)
	}
	return strings.Join(strs, ", ")
}

func exprsEqual(ee1, ee2 []Expr) bool {
	if len(ee1) != len(ee2) {
		return false
	}
	for i := range ee1 {
		if !ee1[i].equal(ee2[i]) {
			return false
		}
	}
	return true
}

func panicf(msg string, args ...interface{}) {
	panic(fmt.Sprintf(msg, args...))
}
