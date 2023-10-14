
use super::gg;
use std::collections::HashMap;

pub enum Error{
    FunctionFromGraph(String),
}

enum ResolvedValue<'a> {
    Value(gg::Value),
    Function(&'a dyn Fn(gg::Value) -> gg::Value),
}

struct Scope<'a> {
    parent: Option<&'a Scope<'a>>,
    values: HashMap<&'a str, ResolvedValue<'a>>,
}

impl<'a> Scope<'a> {

    fn new(parent: Option<&'a Scope<'a>>) -> Scope<'a> {
        Scope {
            parent: parent,
            values: HashMap::new(),
        }
    }

    fn resolve(&self, name: &str) -> Option<&ResolvedValue> {

        if let Some(rval) = self.values.get(name) {
            Some(rval)

        } else if let Some(parent) = self.parent {
            parent.resolve(name)

        } else {
            None
        }
    }

    fn insert(&mut self, name: &'a str, rval: ResolvedValue<'a>) {
        self.values.insert(name, rval);
    }
}

fn always(value: gg::Value) -> Box<dyn Fn(gg::Value) -> gg::Value> {
    Box::new(move |_: gg::Value| value)
}

//fn function_from_graph<'a>(
//    scope: &'a Scope<'a>, g: gg::Graph,
//) -> Result<&'a dyn Fn(gg::Value) -> gg::Value, Error> {
//
//}
