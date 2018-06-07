package gg

import (
	"encoding/json"
	"fmt"
)

type openEdgeJSON struct {
	From    vertexJSON `json:"from"`
	ValueID string     `json:"valueID"`
}

type vertexJSON struct {
	Type    VertexType     `json:"type"`
	ValueID string         `json:"valueID,omitempty"`
	In      []openEdgeJSON `json:"in"`
}

type graphJSON struct {
	Values        map[string]json.RawMessage `json:"values"`
	ValueVertices []vertexJSON               `json:"valueVertices"`
}

// MarshalJSON implements the json.Marshaler interface for a Graph. All Values
// in the Graph will have json.Marshal called on them as-is in order to marshal
// them.
func (g *Graph) MarshalJSON() ([]byte, error) {
	gJ := graphJSON{
		Values:        map[string]json.RawMessage{},
		ValueVertices: make([]vertexJSON, 0, len(g.vM)),
	}

	withVal := func(val Value) (string, error) {
		if _, ok := gJ.Values[val.ID]; !ok {
			valJ, err := json.Marshal(val.V)
			if err != nil {
				return "", err
			}
			gJ.Values[val.ID] = json.RawMessage(valJ)
		}
		return val.ID, nil
	}

	// two locally defined, mutually recursive functions. This kind of thing
	// could probably be abstracted out, I feel like it happens frequently with
	// graph code.
	var mkIns func([]OpenEdge) ([]openEdgeJSON, error)
	var mkVert func(vertex) (vertexJSON, error)

	mkIns = func(in []OpenEdge) ([]openEdgeJSON, error) {
		inJ := make([]openEdgeJSON, len(in))
		for i := range in {
			valID, err := withVal(in[i].val)
			if err != nil {
				return nil, err
			}
			vJ, err := mkVert(in[i].fromV)
			if err != nil {
				return nil, err
			}
			inJ[i] = openEdgeJSON{From: vJ, ValueID: valID}
		}
		return inJ, nil
	}

	mkVert = func(v vertex) (vertexJSON, error) {
		ins, err := mkIns(v.in)
		if err != nil {
			return vertexJSON{}, err
		}
		vJ := vertexJSON{
			Type: v.VertexType,
			In:   ins,
		}
		if v.VertexType == ValueVertex {
			valID, err := withVal(v.val)
			if err != nil {
				return vJ, err
			}
			vJ.ValueID = valID
		}
		return vJ, nil
	}

	for _, v := range g.vM {
		vJ, err := mkVert(v)
		if err != nil {
			return nil, err
		}
		gJ.ValueVertices = append(gJ.ValueVertices, vJ)
	}

	return json.Marshal(gJ)
}

type jsonUnmarshaler struct {
	g  *Graph
	fn func(json.RawMessage) (interface{}, error)
}

// JSONUnmarshaler returns a json.Unmarshaler instance which, when used, will
// unmarshal a json string into the Graph instance being called on here.
//
// The passed in function is used to unmarshal Values (used in both ValueVertex
// vertices and edges) from json strings into go values. The returned inteface{}
// should have already had the unmarshal from the given json string performed on
// it.
//
// The json.Unmarshaler returned can be used many times, but will reset the
// Graph completely before each use.
func (g *Graph) JSONUnmarshaler(fn func(json.RawMessage) (interface{}, error)) json.Unmarshaler {
	return jsonUnmarshaler{g: g, fn: fn}
}

func (jm jsonUnmarshaler) UnmarshalJSON(b []byte) error {
	*(jm.g) = Graph{}
	jm.g.vM = map[string]vertex{}

	var gJ graphJSON
	if err := json.Unmarshal(b, &gJ); err != nil {
		return err
	}

	vals := map[string]Value{}
	getVal := func(valID string) (Value, error) {
		if val, ok := vals[valID]; ok {
			return val, nil
		}

		j, ok := gJ.Values[valID]
		if !ok {
			return Value{}, fmt.Errorf("unmarshaling malformed graph, value with ID %q not defined", valID)
		}

		V, err := jm.fn(j)
		if err != nil {
			return Value{}, err
		}

		val := Value{ID: valID, V: V}
		vals[valID] = val
		return val, nil
	}

	var mkIns func([]openEdgeJSON) ([]OpenEdge, error)
	var mkVert func(vertexJSON) (vertex, error)

	mkIns = func(inJ []openEdgeJSON) ([]OpenEdge, error) {
		in := make([]OpenEdge, len(inJ))
		for i := range inJ {
			val, err := getVal(inJ[i].ValueID)
			if err != nil {
				return nil, err
			}
			v, err := mkVert(inJ[i].From)
			if err != nil {
				return nil, err
			}
			in[i] = OpenEdge{fromV: v, val: val}
		}
		return in, nil
	}

	mkVert = func(vJ vertexJSON) (vertex, error) {
		ins, err := mkIns(vJ.In)
		if err != nil {
			return vertex{}, err
		}
		var val Value
		if vJ.Type == ValueVertex {
			if val, err = getVal(vJ.ValueID); err != nil {
				return vertex{}, err
			}
		}
		return mkVertex(vJ.Type, val, ins...), nil
	}

	for _, v := range gJ.ValueVertices {
		v, err := mkVert(v)
		if err != nil {
			return err
		}
		jm.g.vM[v.id] = v
	}
	return nil
}
