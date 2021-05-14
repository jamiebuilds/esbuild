package html_lexer

import (
	"unicode/utf8"

	"github.com/evanw/esbuild/internal/logger"
)

type T uint8

const eof = -1

const (
	TEndOfFile T = iota
	TSyntaxError

	TDoctypeStart      // "<!doctype"
	TCommentStart      // "<!--"
	TCommentEnd        // "-->"
	TTagOpenStart      // "<"
	TTagCloseStart     // "</"
	TTagEnd            // ">"
	TTagSelfClosingEnd // "/>"
	TEquals            // "="
	TName
	TString
	TText
)

var tokenToString = map[T]string{
	TEndOfFile:         "end of file",
	TSyntaxError:       "syntax error",
	TDoctypeStart:      "<!doctype",
	TCommentStart:      "<!--",
	TCommentEnd:        "-->",
	TTagOpenStart:      "<",
	TTagCloseStart:     "</",
	TTagEnd:            ">",
	TTagSelfClosingEnd: "/>",
	TEquals:            "=",
	TName:              "name",
	TString:            "string",
	TText:              "text",
}

func (t T) String() string {
	return tokenToString[t]
}

// This token struct is designed to be memory-efficient. It just references a
// range in the input file instead of directly containing the substring of text
// since a range takes up less memory than a string.
type Token struct {
	Range logger.Range // 8 bytes
	Kind  T            // 2 bytes
}

type Lexer struct {
	log       logger.Log
	source    logger.Source
	current   int
	codePoint rune
	Token     Token
}

func Tokenize(log logger.Log, source logger.Source) (tokens []Token) {
	lexer := Lexer{
		log:    log,
		source: source,
	}

	lexer.step()
	lexer.next()
	for lexer.Token.Kind != TEndOfFile {
		tokens = append(tokens, lexer.Token)
		lexer.next()
	}
	return
}

func (lexer *Lexer) step() {
	codePoint, width := utf8.DecodeRuneInString(lexer.source.Contents[lexer.current:])

	if width == 0 {
		codePoint = eof
	}

	lexer.codePoint = codePoint
	lexer.Token.Range.Len = int32(lexer.current) - lexer.Token.Range.Loc.Start
	lexer.current += width
}

func (lexer *Lexer) next() {
	lexer.Token = Token{Range: logger.Range{Loc: logger.Loc{Start: lexer.Token.Range.End()}}}
	switch lexer.codePoint {
	case eof:
		lexer.Token.Kind = TEndOfFile
	case '=':
		lexer.step()
		lexer.Token.Kind = TEquals
	case '"', '\'':
		lexer.consumeToEndOfString(lexer.Token.Range, lexer.codePoint)
		lexer.Token.Kind = TString
	case '<':
		if lexer.eat("!doctype") || lexer.eat("!DOCTYPE") {
			lexer.step()
			lexer.Token.Kind = TDoctypeStart
		} else if lexer.eat("!--") {
			lexer.step()
			lexer.Token.Kind = TCommentStart
		} else if lexer.eat("/") {
			lexer.step()
			lexer.Token.Kind = TTagCloseStart
		} else {
			lexer.step()
			if lexer.wouldStartName() {
				lexer.Token.Kind = TTagOpenStart
			} else {
				lexer.consumeToEndOfText(lexer.Token.Range)
				lexer.Token.Kind = TText
			}
		}
	case '-':
		if lexer.eat("->") {
			lexer.step()
			lexer.Token.Kind = TCommentEnd
		} else {
			lexer.step()
			lexer.consumeToEndOfText(lexer.Token.Range)
			lexer.Token.Kind = TText
		}
	case '/':
		if lexer.eat(">") {
			lexer.step()
			lexer.Token.Kind = TTagSelfClosingEnd
		} else {
			lexer.step()
			lexer.consumeToEndOfText(lexer.Token.Range)
			lexer.Token.Kind = TText
		}
	case '>':
		lexer.step()
		lexer.Token.Kind = TTagEnd
	default:
		if lexer.wouldStartName() {
			lexer.consumeToEndOfName(lexer.Token.Range)
			lexer.Token.Kind = TName
		} else {
			lexer.consumeToEndOfText(lexer.Token.Range)
			lexer.Token.Kind = TText
		}
	}
}

func (lexer *Lexer) match(matcher string) bool {
	end := lexer.current + len(matcher)
	return end <= len(lexer.source.Contents) &&
		lexer.source.Contents[lexer.current:end] == matcher
}

func (lexer *Lexer) eat(matcher string) bool {
	if lexer.match(matcher) {
		for i := 0; i < len(matcher); i++ {
			lexer.step()
		}
		return true
	} else {
		return false
	}
}

func (lexer *Lexer) consumeToEndOfString(startRange logger.Range, startChar rune) {
	for {
		switch lexer.codePoint {
		case eof:
			lexer.log.AddRangeError(&lexer.source, lexer.Token.Range, "String must be closed")
			return
		case startChar:
			lexer.step()
			return
		default:
			lexer.step()
		}
	}
}

func (lexer *Lexer) consumeToEndOfText(startRange logger.Range) {
	for {
		switch lexer.codePoint {
		case '<', '-', '>', '"', '\'', eof:
			return
		default:
			lexer.step()
		}
	}
}

func IsNameStart(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c >= 0x80
}

func IsNameContinue(c rune) bool {
	return IsNameStart(c) || c == ':'
}

func (lexer *Lexer) wouldStartName() bool {
	return IsNameStart(lexer.codePoint)
}

func (lexer *Lexer) consumeToEndOfName(startRange logger.Range) {
	for IsNameContinue(lexer.codePoint) {
		lexer.step()
	}
}
