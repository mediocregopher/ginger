# Graphs

## Definitions

- All values are immutable
- All values have a type. These are the types necessary to construct and
  traverse graphs:

    - Tuple
        - Contains zero or more member values, each possibly of different types
        - Size is finite and known upon creation

    - Iterator
        - Produces zero or more values in sequence, each being of the same type

    - Bool
        - Binary value, true or false

    - Graph
        - Unordered set of edges

        - Edge
            - Identified by a 3-tuple of (Vertex, $val, Vertex), with each tuple
              being unique within a graph

        - Vertex
            - Has ordered set of in edges
            - Has ordered set of out edges
            - 3 types of vertices
                - Node
                    - Contains exactly one value of any type
                    - Is unique within graph, based on its value
                    - Has at least one edge (either in or out)
                - Junction
                    - Has two or more in edges
                    - Has exactly one out edge
                - Fork
                    - Has exactly one in edge
                    - Has one or more out edges

        - Half-edge
            - Has no properties, simply exists

### Example graph

Here is a graph which will, when interpreted and compiled in a certain way,
take the average of an input tuple containing only integers:

```
            +    \
        |------- |
        |        |  /
 in --- |        |---- out
        |  size  |
        |------- |
                 /
```

- The first "wall" of pipe characters represents a fork, where the input is
  copied into two different edges. In the top edge the elements of the input
  tuple are summed using the `+` attribute. In the bottom edge the
  number of elements in the input tuple are counted using the `size` attribute.

- The second "wall" of pipe characters represents a junction (note the top and
  bottom slashes to differentiate from a fork). A junction combines its input
  edges into a new tuple. So this junction creates a 2-tuple, the first element
  being the sum of the `in` tuple's elements, the second being the count of how
  many elements there were in `in`.

- Finally, the members of that 2-tuple are divided using the `/` attribute on
  the edge leaving the junction. As this edge is the input into the `out` node,
  the result of the division becomes the output of this graph.

## Operations

* Each operation has exactly one input value and one output value, and specifies
  the type of each.
* A string preceded by `$` indicates any value of any type. The string gives
  context as to how that value will be used.
* Abbreviations
    - `($T0,...,$Tn)`: Tuple containing zero or more members of varying types
    - `($T...)` : Tuple containing zero or more members of the same type
    - `V`: vertex of any type (node, junction, or fork)
    - `E`: edge
    - `e`: half-edge
    - `G`: graph
    - `it<$T>`: iterator whose values are all type `$T`

```
# Operation syntax:
# name : input -> output

# Tuple and iterator basics
tup_it  : ($T...) -> it<$T>
it_next : it<$T> -> (it<$T>, $T, bool)

# Graph construction
graph_mk_edge_from : ($val, $attr) -> e
graph_mk_edge_to   : (e, $val) -> G
graph_mk_fork      : (e, $attr) -> e
graph_mk_junction  : (it<e>, $attr) -> e
graph_merge        : (G, G) -> G

# Graph traversal
graph_edge_in         : E -> V
graph_edge_out        : E -> V
graph_edge_attr       : E -> $attr
graph_vertex_ins      : V -> it<E>
graph_vertex_outs     : V -> it<E>
graph_vertex_node_val : V -> ($val, bool)
graph_nodes           : G -> it<V>
graph_node            : (G,$val) -> (V, bool) # maybe not needed?
```

### Example graph construction

The graph above could be constructed in the following way:

```
eIn    = graph_mk_edge_from(in, ())
eSum   = graph_mk_fork(eIn, +)
eCount = graph_mk_fork(eIn, size)
eAvg   = graph_mk_junction(tup_it(eSum, eCount), /)
G      = graph_mk_edge_to(eAvg, out)

return G
```

## Notes

- Assuming _only_ the given operations are used:
    - It is impossible to construct an invalid graph.
    - It is impossible to call any graph traversal operations in an invalid or
      undefined way.
