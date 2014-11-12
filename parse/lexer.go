package parse

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// token represents a tokenType or text string returned from the scanner.
type token struct {
	typ tokenType // The type of this token.
	pos Pos       // The starting position, in bytes, of this token in the input string.
	val string    // The value of this token.
}

func (i token) String() string {
	switch {
	case i.typ == tokenEOF:
		return "EOF"
	case i.typ == tokenError:
		return i.val
	case i.typ > tokenKeyword:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// tokenType identifies the type of lex items.
type tokenType int

const (
	tokenError tokenType = iota // error occurred; value is text of error
	tokenEOF
	tokenSpace
	tokenIdentifier
	tokenStar
	tokenSemiColon
	tokenDataType
	tokenComma
	tokenLeftParentheses
	tokenRightParentheses
	tokenString
	tokenNumber
	//keywords
	tokenKeyword
	tokenSelect
	tokenFrom
	tokenCreate
	tokenTable
	tokenInsert
	tokenInto
	tokenValues
)

var keys = map[string]tokenType{
	"SELECT": tokenSelect,
	"FROM":   tokenFrom,
	"CREATE": tokenCreate,
	"TABLE":  tokenTable,
	"INSERT": tokenInsert,
	"INTO":   tokenInto,
	"VALUES": tokenValues,
}

var dataTypes = map[string]tokenType{
	"int":  tokenDataType,
	"text": tokenDataType,
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input      string     // the string being scanned
	state      stateFn    // the next lexing function to enter
	pos        Pos        // current position in the input
	start      Pos        // start position of this token
	width      Pos        // width of last rune read from input
	lastPos    Pos        // position of most recent token returned by nextItem
	items      chan token // channel of scanned items
	parenDepth int        // nesting depth of ( ) exprs
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an token back to the client.
func (l *lexer) emit(t tokenType) {
	i := token{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
	l.items <- i
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous token returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error tokenType and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- token{tokenError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next token from the input.
func (l *lexer) nextItem() token {
	token := <-l.items
	l.lastPos = token.pos
	return token
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan token),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexText; l.state != nil; {
		l.state = l.state(l)
	}
}

func lexText(l *lexer) stateFn {
	switch r := l.next(); {
	case isSpace(r):
		return lexSpace
	case r == ';':
		l.emit(tokenSemiColon)
	case r == '*':
		l.emit(tokenStar)
	case r == '+' || r == '-' || ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case isAlphaNumeric(r):
		return lexIdentifier
	case r == '(':
		l.emit(tokenLeftParentheses)
	case r == ')':
		l.emit(tokenRightParentheses)
	case r == ',':
		l.emit(tokenComma)
	case r == eof:
		l.emit(tokenEOF)
		return nil
	case r == '\'', r == '"':
		return lexString
	}
	return lexText
}

func lexString(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
		// absorb.
		case r == '\'', r == '"':
			l.emit(tokenString)
			break Loop
		}
	}
	return lexText
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(tokenSpace)
	return lexText
}

// lexIdentifier scans an alphanumeric.
func lexIdentifier(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			switch {
			case keys[word] > tokenKeyword:
				l.emit(keys[word])
			case dataTypes[word] > 0:
				l.emit(dataTypes[word])
			default:
				l.emit(tokenIdentifier)
			}
			break Loop
		}
	}
	return lexText
}

// lexNumber scans a number: decimal, octal, hex, float, or imaginary. This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(tokenNumber)
	return lexText
}

func (l *lexer) scanNumber() bool {
	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	// Is it imaginary?
	l.accept("i")
	// Next thing mustn't be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return false
	}
	return true
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
