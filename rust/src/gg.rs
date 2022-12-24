pub mod lexer;

use super::graph::Graph;

#[derive(Clone, Eq, Hash, PartialEq, Debug)]
pub enum Value<'a>{
    Name(&'a str),
    Number(i64),
    Graph(&'a Graph<Value<'a>, Value<'a>>),
}
