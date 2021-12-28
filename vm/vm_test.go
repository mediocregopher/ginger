package vm

import (
	"bytes"
	"testing"

	"github.com/mediocregopher/ginger/gg"
	"github.com/stretchr/testify/assert"
)

func TestVM(t *testing.T) {

	src := `
		incr = { out = add < (1; in;); };

		out = incr < incr < in;
	`

	var in int64 = 5

	val, err := EvaluateSource(
		bytes.NewBufferString(src),
		gg.Value{Number: &in},
		GlobalScope,
	)

	assert.NoError(t, err)
	assert.Equal(t, in+2, *val.Number)
}
