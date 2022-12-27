use std::hash::{Hash, Hasher};
use im_rc::{HashSet};

mod lexer;
mod decoder;

pub use decoder::Decoder;

#[derive(Clone, Hash, PartialEq, Eq, PartialOrd, Ord)]
#[cfg_attr(test, derive(Debug))]
pub struct OpenEdge(Value, Value); // edge, src

#[derive(Clone, Hash, PartialEq, Eq, PartialOrd, Ord)]
#[cfg_attr(test, derive(Debug))]
pub enum Value{
    Name(String),
    Number(i64),
    Tuple(Vec<OpenEdge>),
    Graph(Graph),
}

pub const ZERO_TUPLE: Value = Value::Tuple(vec![]);

#[derive(Clone, PartialEq, Eq, PartialOrd, Ord)]
#[cfg_attr(test, derive(Debug))]
pub struct Graph {
    edges: HashSet<(Value, OpenEdge)>, // dst, src
}

impl Graph {

    pub fn new() -> Graph {
        Graph{edges: HashSet::new()}
    }

    pub fn with(&self, dst_val: Value, edge_val: Value, src_val: Value) -> Self {
        Graph{
            edges: self.edges.update((dst_val, OpenEdge(edge_val, src_val))),
        }
    }
}

// The implementation of hash for im_rc::HashSet does not sort the entries.
impl Hash for Graph {
    fn hash<H: Hasher>(&self, state: &mut H) {
        let mut edges = Vec::from_iter(&self.edges);
        edges.sort();
        edges.iter().for_each(|edge| edge.hash(state));
    }
}

#[cfg(test)]
mod tests {

    use super::*;

    fn number(i: i64) -> Value {
        Value::Number(i)
    }

    #[test]
    fn equality() {

        let g1 = Graph::new()
            .with(number(0), number(1), number(2))
            .with(number(3), number(4), number(5));

        let g2 = Graph::new()
            .with(number(3), number(4), number(5))
            .with(number(0), number(1), number(2));

        assert_eq!(g1, g2);
    }

    #[test]
    fn deep_equality() {

        let g1 = Graph::new().with(number(-2), ZERO_TUPLE, Value::Graph(Graph::new()
            .with(number(0), number(1), number(2))
            .with(number(3), number(4), number(5)),
        ));

        let g2 = Graph::new().with(number(-2), ZERO_TUPLE, Value::Graph(Graph::new()
            .with(number(3), number(4), number(5))
            .with(number(0), number(1), number(2)),
        ));

        assert_eq!(g1, g2);
    }
}
