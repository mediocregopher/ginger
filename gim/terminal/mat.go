package terminal

import (
	"container/list"
)

type matEl struct {
	x int
	v interface{}
}

type matRow struct {
	y int
	l *list.List
}

// a 2-d sparse matrix
type mat struct {
	rows *list.List

	currY     int
	currRowEl *list.Element
	currEl    *list.Element
}

func newMat() *mat {
	return &mat{
		rows: list.New(),
	}
}

func (m *mat) getRow(y int) *list.List {
	m.currY = y             // this will end up being true no matter what
	if m.currRowEl == nil { // first call
		l := list.New()
		m.currRowEl = m.rows.PushFront(matRow{y: y, l: l})
		return l

	} else if m.currRowEl.Value.(matRow).y > y {
		m.currRowEl = m.rows.Front()
	}

	for {
		currRow := m.currRowEl.Value.(matRow)
		switch {
		case currRow.y == y:
			return currRow.l
		case currRow.y < y:
			if m.currRowEl = m.currRowEl.Next(); m.currRowEl == nil {
				l := list.New()
				m.currRowEl = m.rows.PushBack(matRow{y: y, l: l})
				return l
			}
		default: // currRow.y > y
			l := list.New()
			m.currRowEl = m.rows.InsertBefore(matRow{y: y, l: l}, m.currRowEl)
			return l
		}
	}
}

func (m *mat) getEl(x, y int) *matEl {
	var rowL *list.List
	if m.currRowEl == nil || m.currY != y {
		rowL = m.getRow(y)
		m.currEl = rowL.Front()
	} else {
		rowL = m.currRowEl.Value.(matRow).l
	}

	if m.currEl == nil || m.currEl.Value.(*matEl).x > x {
		if m.currEl = rowL.Front(); m.currEl == nil {
			// row is empty
			mel := &matEl{x: x}
			m.currEl = rowL.PushFront(mel)
			return mel
		}
	}

	for {
		currEl := m.currEl.Value.(*matEl)
		switch {
		case currEl.x == x:
			return currEl
		case currEl.x < x:
			if m.currEl = m.currEl.Next(); m.currEl == nil {
				mel := &matEl{x: x}
				m.currEl = rowL.PushBack(mel)
				return mel
			}
		default: // currEl.x > x
			mel := &matEl{x: x}
			m.currEl = rowL.InsertBefore(mel, m.currEl)
			return mel
		}
	}
}

func (m *mat) get(x, y int) interface{} {
	return m.getEl(x, y).v
}

func (m *mat) set(x, y int, v interface{}) {
	m.getEl(x, y).v = v
}

func (m *mat) iter(f func(x, y int, v interface{}) bool) {
	for rowEl := m.rows.Front(); rowEl != nil; rowEl = rowEl.Next() {
		row := rowEl.Value.(matRow)
		for el := row.l.Front(); el != nil; el = el.Next() {
			mel := el.Value.(*matEl)
			if !f(mel.x, row.y, mel.v) {
				return
			}
		}
	}
}
