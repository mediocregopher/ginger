use std::hash::Hash;
use im_rc::{HashMap,HashSet};

#[derive(Clone, Eq, Hash, PartialEq, Debug)]
pub enum OpenEdgeSource<E, V>{
    Value(V),
    Tuple(Vec<OpenEdge<E, V>>),
}

#[derive(Clone, Eq, Hash, PartialEq, Debug)]
pub struct OpenEdge<E, V>{
    value: E,
    source: OpenEdgeSource<E, V>,
}

#[derive(Clone, Eq, Hash, PartialEq, Debug)]
pub struct Graph<E, V>
where
    E: Hash + Eq + Clone,
    V: Hash + Eq + Clone,
{
    roots: HashMap<V, HashSet<OpenEdge<E, V>>>,
}

impl<E, V> Graph<E, V>
where
    E: Hash + Eq + Clone,
    V: Hash + Eq + Clone,
{

    pub fn new() -> Graph<E, V> {
        Graph{roots: HashMap::new()}
    }

    pub fn from_value(edge_value: E, source_value: V) -> OpenEdge<E, V> {
        OpenEdge{
            value: edge_value,
            source: OpenEdgeSource::Value(source_value),
        }
    }

    pub fn from_tuple(edge_value: E, source_tuple: Vec<OpenEdge<E, V>>) -> OpenEdge<E, V> {
        OpenEdge{
            value: edge_value,
            source: OpenEdgeSource::Tuple(source_tuple),
        }
    }

    pub fn with(&self, root_value: V, open_edge: OpenEdge<E, V>) -> Self {

        let new_roots = self.roots.alter(
            |set_option: Option<HashSet<OpenEdge<E, V>>>| -> Option<HashSet<OpenEdge<E, V>>> {
                match set_option {
                    None => Some(HashSet::unit(open_edge)),
                    Some(set) => Some(set.update(open_edge)),
                }
            },
            root_value,
        );

        Graph{
            roots: new_roots,
        }
    }
}

