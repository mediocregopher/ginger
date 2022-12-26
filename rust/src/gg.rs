use std::hash::Hash;
use im_rc::{HashSet};

pub mod lexer;

#[derive(Clone, Eq, Hash, PartialEq)]
pub struct OpenEdge(Value, Value); // edge, src

#[derive(Clone, Eq, Hash, PartialEq)]
pub enum Value{
    Name(String),
    Number(i64),
    Tuple(Vec<OpenEdge>),
    Graph(Graph),
}

#[derive(Clone, Eq, Hash, PartialEq)]
struct Edge<V> {
    dst_val: V,
    src_val: V,
}

#[derive(Clone, Eq, Hash, PartialEq)]
pub struct Graph {
    edges: HashSet<(Value, OpenEdge)>, // dst, src
}

impl Graph {

    pub fn new() -> Graph {
        Graph{edges: HashSet::new()}
    }

    pub fn with(&self, dst_val: Value, src_edge: OpenEdge) -> Self {
        Graph{
            edges: self.edges.update((dst_val, src_edge)),
        }
    }
}
