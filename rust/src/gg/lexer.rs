use std::fmt;
use std::io::{self, Read, BufReader};
use unicode_categories::UnicodeCategories;

use char_reader::CharReader;

#[derive(Copy, Clone, Debug, PartialEq)]
pub struct Location {
    pub row: i64,
    pub col: i64,
}

impl fmt::Display for Location {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}:{}", self.row, self.col)
    }
}

#[derive(Debug)]
pub enum Error {
    Tokenizing(&'static str, Location),
    IO(io::Error),
}

impl From<io::Error> for Error{
    fn from(e: io::Error) -> Self {
        Error::IO(e)
    }
}

#[derive(Debug, PartialEq)]
pub enum TokenKind {
    Name,
    Number,
    Punctuation,
}

#[derive(Debug, PartialEq)]
pub struct Token {
    pub kind: TokenKind,
    pub value: String,
}

pub struct Lexer<R: Read> {
    r: CharReader<BufReader<R>>,
    buf: String,

    next_loc: Location,
}

impl<R: Read> Lexer<R>{

    pub fn new(r: R) -> Lexer<R> {
        Lexer{
            r: CharReader::new(BufReader::new(r)),
            buf: String::new(),
            next_loc: Location{
                row: 0,
                col: 0,
            },
        }
    }

    fn discard(&mut self) {

        let c = self.r.next_char().
            expect("discard should only get called after peek").
            expect("discard should only get called after peek");

        if c == '\n' {
            self.next_loc = Location{
                row: self.next_loc.row + 1,
                col: 0
            };
            return;
        }

        self.next_loc = Location{
            row: self.next_loc.row,
            col: self.next_loc.col + 1,
        };
    }

    fn peek_a_bool(&mut self) -> Result<(char, bool), Error> {
        if let Some(c) = self.r.peek_char()? {
            Ok((c, true))
        } else {
            Ok(('0', false))
        }
    }

    fn discard_while(&mut self, pred: impl Fn(char) -> bool) -> Result<(), Error> {

        loop {
            let (c, ok) = self.peek_a_bool()?;
            if !ok || !pred(c) {
                return Ok(());
            }

            self.discard();
        }
    }

    fn collect_token(
        &mut self,
        kind: TokenKind,
        pred: impl Fn(char) -> bool,
    ) -> Result<Option<(Token, Location)>, Error> {

        let loc = self.next_loc;
        self.buf.truncate(0);

        loop {

            let (c, ok) = self.peek_a_bool()?;

            if !ok || !pred(c) {
                return Ok(Some(
                    (Token{kind: kind, value: self.buf.clone()}, loc)
                ))
            }

            self.buf.push(c);
            self.discard();
        }
    }

    fn is_number(c: char) -> bool {
        c == '-' || ('0' <= c && c <= '9')
    }

    pub fn next(&mut self) -> Result<Option<(Token, Location)>, Error> {

        loop {

            let (c, ok) = self.peek_a_bool()?;
            if !ok {
                return Ok(None);

            } else if c == '*' {
                self.discard_while(|c| c != '\n')?;
                // the terminating newline will be dealt with in the next loop

            } else if c.is_letter() {
                return self.collect_token(
                    TokenKind::Name,
                    |c| c.is_letter() || c.is_number() || c.is_mark() || c == '-',
                );

            } else if Self::is_number(c) {
                return self.collect_token(TokenKind::Number, Self::is_number);

            } else if c.is_punctuation() || c.is_symbol() {

                let loc = self.next_loc;
                self.discard();

                return Ok(Some(
                    (Token{kind: TokenKind::Punctuation, value: c.to_string()}, loc)
                ))

            } else if c.is_ascii_whitespace() {
                self.discard_while(|c| c.is_ascii_whitespace())?;

            } else {
                return Err(Error::Tokenizing("unexpected character", self.next_loc));
            }
        }
    }
}

#[cfg(test)]
mod tests {

    use super::*;

    #[test]
    fn lexer() {

        struct Test {
            input: &'static str,
            exp: Vec<(Token, Location)>,
        }

        fn tok(kind: TokenKind, val: &'static str, loc_row: i64, loc_col: i64) -> (Token, Location) {
            (
                Token{kind: kind, value: val.to_string()},
                Location{row: loc_row, col: loc_col},
            )
        }

        let tests = vec![
            Test{
                input: "",
                exp: vec![],
            },
            Test{
                input: "* foo",
                exp: vec![],
            },
            Test{
                input: "* foo\n",
                exp: vec![],
            },
            Test{
                input: "* foo\nbar",
                exp: vec![
                    tok(TokenKind::Name, "bar", 1, 0),
                ],
            },
            Test{
                input: "* foo\nbar ",
                exp: vec![
                    tok(TokenKind::Name, "bar", 1, 0),
                ],
            },
            Test{
                input: "foo bar\nf-o f0O Foo",
                exp: vec![
                    tok(TokenKind::Name, "foo", 0, 0),
                    tok(TokenKind::Name, "bar", 0, 4),
                    tok(TokenKind::Name, "f-o", 1, 0),
                    tok(TokenKind::Name, "f0O", 1, 4),
                    tok(TokenKind::Name, "Foo", 1, 8),
                ],
            },
            Test{
                input: "1 100 -100",
                exp: vec![
                    tok(TokenKind::Number, "1", 0, 0),
                    tok(TokenKind::Number, "100", 0, 2),
                    tok(TokenKind::Number, "-100", 0, 6),
                ],
            },
            Test{
                input: "1<2!-3 ()",
                exp: vec![
                    tok(TokenKind::Number, "1", 0, 0),
                    tok(TokenKind::Punctuation, "<", 0, 1),
                    tok(TokenKind::Number, "2", 0, 2),
                    tok(TokenKind::Punctuation, "!", 0, 3),
                    tok(TokenKind::Number, "-3", 0, 4),
                    tok(TokenKind::Punctuation, "(", 0, 7),
                    tok(TokenKind::Punctuation, ")", 0, 8),
                ],
            },
        ];

        for test in tests {
            println!("INPUT: {:#?}", test.input);

            let mut l = Lexer::new(test.input.as_bytes());
            let mut res = Vec::new();

            loop {
                if let Some(token) = l.next().expect("no errors expected") {
                    res.push(token);
                } else {
                    break;
                }
            }

            assert_eq!(*res.as_slice(), *test.exp)
        }
    }
}

