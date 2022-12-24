use std::fmt;
use std::io::{self, Read, BufReader};
use unicode_categories::UnicodeCategories;

use char_reader::CharReader;

pub struct Location {
    pub row: i64,
    pub col: i64,
}

impl fmt::Display for Location {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}:{}", self.row, self.col)
    }
}

pub enum Error {
    Tokenizing(&'static str, Location),
    IO(io::Error),
}

impl From<io::Error> for Error{
    fn from(e: io::Error) -> Self {
        Error::IO(e)
    }
}

pub enum TokenKind {
    Name,
    Number,
    Punctuation,
}

pub struct Token {
    pub kind: TokenKind,
    pub value: String,
    pub location: Location,
}

pub struct Lexer<R: Read> {
    r: CharReader<BufReader<R>>,
    buf: String,

    prev_char: char,
    prev_loc: Location,
}

impl<R: Read> Lexer<R>{

    fn next_loc(&self) -> Location {

        if self.prev_char == '\n' {
            return Location{
                row: self.prev_loc.row + 1,
                col: 0
            };
        }

        return Location{
            row: self.prev_loc.row,
            col: self.prev_loc.col + 1,
        }
    }

    fn discard(&mut self) {

        self.prev_char = self.r.next_char().
            expect("discard should only get called after peek").
            expect("discard should only get called after peek");

        self.prev_loc = self.next_loc();
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
    ) -> Result<Option<Token>, Error> {

        let loc = self.next_loc();
        self.buf.truncate(0);

        loop {

            let (c, ok) = self.peek_a_bool()?;

            if !ok || !pred(c) {
                return Ok(Some(Token{
                    kind: kind,
                    value: self.buf.clone(),
                    location: loc,
                }))
            }

            self.buf.push(c);
            self.discard();
        }
    }

    fn is_number(c: char) -> bool {
        c == '-' || ('0' <= c && c <= '9')
    }

    pub fn next(&mut self) -> Result<Option<Token>, Error> {

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

            } else if c.is_punctuation() {
                return self.collect_token(
                    TokenKind::Punctuation,
                    |c| c.is_punctuation() || c.is_symbol(),
                );

            } else if c.is_ascii_whitespace() {
                self.discard_while(|c| c.is_ascii_whitespace())?;

            } else {
                return Err(Error::Tokenizing("unexpected character", self.next_loc()));
            }
        }
    }
}
