package gg

import "io"

type mockReader struct {
	body []byte
	err  error
}

func (r *mockReader) Read(b []byte) (int, error) {

	n := copy(b, r.body)
	r.body = r.body[n:]

	if len(r.body) == 0 {
		if r.err == nil {
			return n, io.EOF
		}
		return n, r.err
	}

	return n, nil
}
