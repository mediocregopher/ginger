use std::io::{self, Read};

use super::{Graph, Value, OpenEdge, ZERO_TUPLE};
use super::lexer::{self, Lexer, Token, TokenKind, Location};

// In order to make sense of this file, check out the accompanying gg.bnf, which describes the
// grammar in BNF notation. Each method in the Decoder maps more or less exactly to a state within
// the BNF.

#[cfg_attr(test, derive(Debug))]
pub enum Error {
    Decoding(String, Location),
    IO(io::Error),
}

impl From<lexer::Error> for Error {
    fn from(e: lexer::Error) -> Self {
        match e {
            lexer::Error::Tokenizing(s, loc) => Error::Decoding(s, loc),
            lexer::Error::IO(e) => Error::IO(e)
        }
    }
}

static OUTER_GRAPH_TERM: Token = Token{
    kind: TokenKind::End,
    value: String::new(),
};

pub struct Decoder<R: Read> {
    lexer: Lexer<R>,
}

impl<R: Read> Decoder<R> {

    pub fn new(r: R) -> Decoder<R> {
        Decoder{
            lexer: Lexer::new(r),
        }
    }

    pub fn decode_undelimited(&mut self) -> Result<Graph, Error> {
        self.outer_graph(Graph::new())
    }

    fn exp_punct(&mut self, v: &'static str) -> Result<(), Error> {

        match self.lexer.next()? {
            (Token{kind: TokenKind::Punctuation, value: v2}, _) if v == v2 => Ok(()),
            (tok, loc) => Err(Error::Decoding(
                format!("expected '{}', found: {}", v, tok),
                loc,
            )),
        }
    }

    fn generic_graph(&mut self, term_tok: &Token, g: Graph) -> Result<Graph, Error> {

        match self.lexer.next()? {

            (tok, _) if tok == *term_tok => Ok(g),

            (Token{kind: TokenKind::Name, value: name}, _) => {
                self.exp_punct("=")?;
                let open_edge = self.generic_graph_tail(term_tok, ZERO_TUPLE)?;
                self.generic_graph(term_tok, g.with(
                    Value::Name(name),
                    open_edge.0,
                    open_edge.1,
                ))
            }

            (tok, loc) => Err(Error::Decoding(
                format!("expected name or {}, found: {}", term_tok, tok),
                loc,
            )),
        }
    }

    fn generic_graph_tail(&mut self, term_tok: &Token, edge_val: Value) -> Result<OpenEdge, Error> {

        let val = self.value()?;

        match self.lexer.next()? {

            (Token{kind: TokenKind::Punctuation, value: v}, _) if v == ";" =>
                Ok(OpenEdge(edge_val, val)),

            (Token{kind: TokenKind::Punctuation, value: v}, _) if v == "<" =>

                if edge_val == ZERO_TUPLE {
                    self.generic_graph_tail(term_tok, val)
                } else {
                    Ok(OpenEdge(edge_val, Value::Tuple(vec![
                        self.generic_graph_tail(term_tok, val)?,
                    ])))
                },

            (tok, loc) => {
                self.lexer.push_next(tok, loc);
                Ok(OpenEdge(edge_val, val))
            },
        }
    }

    fn outer_graph(&mut self, g: Graph) -> Result<Graph, Error> {
        self.generic_graph(&OUTER_GRAPH_TERM, g)
    }

    fn graph(&mut self, g: Graph) -> Result<Graph, Error> {

        let term_tok = Token{
            kind: TokenKind::Punctuation,
            value: String::from("}"),
        };

        self.generic_graph(&term_tok, g)
    }

    fn tuple(&mut self, tuple_vec: &mut Vec<OpenEdge>) -> Result<(), Error> {

        loop {
            match self.lexer.next()? {

                (Token{kind: TokenKind::Punctuation, value: v}, _) if v == ")" =>
                    return Ok(()),

                (tok, loc) => {
                    self.lexer.push_next(tok, loc);
                    tuple_vec.push(self.tuple_tail(ZERO_TUPLE)?);
                },
            }
        }
    }

    fn tuple_tail(&mut self, edge_val: Value) -> Result<OpenEdge, Error> {

        let val = self.value()?;

        match self.lexer.next()? {

            (Token{kind: TokenKind::Punctuation, value: v}, _) if v == "," =>
                Ok(OpenEdge(edge_val, val)),

            (Token{kind: TokenKind::Punctuation, value: v}, _) if v == "<" =>

                if edge_val == ZERO_TUPLE {
                    self.tuple_tail(val)
                } else {
                    Ok(OpenEdge(edge_val, Value::Tuple(vec![
                        self.tuple_tail(val)?,
                    ])))
                },

            (tok, loc) => {
                self.lexer.push_next(tok, loc);
                Ok(OpenEdge(edge_val, val))
            },

        }
    }

    fn value(&mut self) -> Result<Value, Error> {

        match self.lexer.next()? {

            (Token{kind: TokenKind::Name, value: v}, _) =>
                Ok(Value::Name(v)),

            (Token{kind: TokenKind::Number, value: v}, loc) =>
                match v.parse::<i64>() {
                    Ok(n) => Ok(Value::Number(n)),
                    Err(e) => Err(Error::Decoding(
                        format!("parsing {:#?} as integer: {}", v, e),
                        loc,
                    )),
                },

            (Token{kind: TokenKind::Punctuation, value: v}, _) if v == "(" => {
                let mut vec = Vec::new();
                self.tuple(&mut vec)?;
                Ok(Value::Tuple(vec))
            },

            (Token{kind: TokenKind::Punctuation, value: v}, _) if v == "{" =>
                Ok(Value::Graph(self.graph(Graph::new())?)),

            (tok, loc) => Err(Error::Decoding(
                format!("expected name, number, '(', or '{{', found: {}", tok),
                loc,
            )),
        }
    }
}

#[cfg(test)]
mod tests {

    use super::*;

    #[test]
    fn decoder() {

        fn name(s: &'static str) -> Value {
            Value::Name(s.to_string())
        }

        fn number(i: i64) -> Value {
            Value::Number(i)
        }

        struct Test {
            input: &'static str,
            exp: Graph,
        }

        let tests = vec!{
            Test{
                input: "",
                exp: Graph::new(),
            },
            Test{
                input: "out = 1",
                exp: Graph::new().
                        with(name("out"), ZERO_TUPLE, number(1)),
            },
            Test{
                input: "out = 1;",
                exp: Graph::new().
                        with(name("out"), ZERO_TUPLE, number(1)),
            },
            Test{
                input: "out = incr < 1",
                exp: Graph::new().
                        with(name("out"), name("incr"), number(1)),
            },
            Test{
                input: "out = incr < 1;",
                exp: Graph::new().
                        with(name("out"), name("incr"), number(1)),
            },
            Test{
                input: "out = a < b < 1",
                exp: Graph::new().with(
                    name("out"),
                    name("a"),
                    Value::Tuple(vec![OpenEdge(name("b"), number(1))]),
                ),
            },
            Test{
                input: "out = a < b < 1;",
                exp: Graph::new().with(
                    name("out"),
                    name("a"),
                    Value::Tuple(vec![OpenEdge(name("b"), number(1))]),
                ),
            },
            Test{
                input: "out = a < b < (1, c < 2, d < e < 3)",
                exp: Graph::new().with(
                    name("out"),
                    name("a"),
                    Value::Tuple(vec![
                        OpenEdge(name("b"), Value::Tuple(vec![
                            OpenEdge(ZERO_TUPLE, number(1)),
                            OpenEdge(name("c"), number(2)),
                            OpenEdge(name("d"), Value::Tuple(vec![
                                OpenEdge(name("e"), number(3)),
                            ])),
                        ])),
                    ]),
                ),
            },
            Test{
                input: "out = (c < 2,);",
                exp: Graph::new().with(
                    name("out"),
                    ZERO_TUPLE,
                    Value::Tuple(vec![
                        OpenEdge(name("c"), number(2)),
                    ]),
                ),
            },
            Test{
                input: "out = (1, c < 2) < 3;",
                exp: Graph::new().with(
                    name("out"),
                    Value::Tuple(vec![
                        OpenEdge(ZERO_TUPLE, number(1)),
                        OpenEdge(name("c"), number(2)),
                    ]),
                    number(3),
                ),
            },
            Test{
                input: "out = a < b < (1, c < (d < 2, 3))",
                exp: Graph::new().with(
                    name("out"),
                    name("a"),
                    Value::Tuple(vec![
                        OpenEdge(name("b"), Value::Tuple(vec![
                            OpenEdge(ZERO_TUPLE, number(1)),
                            OpenEdge(name("c"), Value::Tuple(vec![
                                OpenEdge(name("d"), number(2)),
                                OpenEdge(ZERO_TUPLE, number(3)),
                            ])),
                        ])),
                    ]),
                ),
            },
            Test{
                input: "out = { a = 1; b = 2 < 3; c = 4 < 5 < 6 }",
                exp: Graph::new().with(
                    name("out"),
                    ZERO_TUPLE,
                    Value::Graph(Graph::new()
                        .with(name("a"), ZERO_TUPLE, number(1))
                        .with(name("b"), number(2), number(3))
                        .with(name("c"), number(4), Value::Tuple(vec![
                            OpenEdge(number(5), number(6)),
                        ])),
                    ),
                ),
            },
            Test{
                input: "out = { a = 1; };",
                exp: Graph::new().with(
                    name("out"),
                    ZERO_TUPLE,
                    Value::Graph(Graph::new()
                        .with(name("a"), ZERO_TUPLE, number(1)),
                    ),
                ),
            },
            Test{
                input: "out = { a = 1; } < 2",
                exp: Graph::new().with(
                    name("out"),
                    Value::Graph(Graph::new()
                        .with(name("a"), ZERO_TUPLE, number(1)),
                    ),
                    number(2),
                ),
            },
            Test{
                input: "out = { a = 1; } < 2; foo = 5 < 6",
                exp: Graph::new()
                    .with(
                        name("out"),
                        Value::Graph(Graph::new()
                            .with(name("a"), ZERO_TUPLE, number(1)),
                        ),
                        number(2),
                    )
                    .with(name("foo"), number(5), number(6)),
            },
            Test{
                input: "out = { a = 1 } < 2; foo = 5 < 6;",
                exp: Graph::new()
                    .with(
                        name("out"),
                        Value::Graph(Graph::new()
                            .with(name("a"), ZERO_TUPLE, number(1)),
                        ),
                        number(2),
                    )
                    .with(name("foo"), number(5), number(6)),
            },
        };

        for test in tests {
            println!("INPUT: {:#?}", test.input);

            let mut d = Decoder::new(test.input.as_bytes());
            let got = d.decode_undelimited().expect("no errors expected");
            assert_eq!(test.exp, got);
        }

    }
}
