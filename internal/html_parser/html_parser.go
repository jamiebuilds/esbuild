package html_parser

import (
	"fmt"

	"github.com/evanw/esbuild/internal/ast"
	"github.com/evanw/esbuild/internal/html_ast"
	"github.com/evanw/esbuild/internal/html_lexer"
	"github.com/evanw/esbuild/internal/logger"
)

type parser struct {
	log           logger.Log
	source        logger.Source
	tokens        []html_lexer.Token
	stack         []html_lexer.T
	index         int
	end           int
	prevError     logger.Loc
	importRecords []ast.ImportRecord
}

func Parse(log logger.Log, source logger.Source) html_ast.AST {
	p := parser{
		log:       log,
		source:    source,
		tokens:    html_lexer.Tokenize(log, source),
		prevError: logger.Loc{Start: -1},
	}
	p.end = len(p.tokens)
	tree := html_ast.AST{}
	tree.Doctype = p.parseDoctype()
	return tree
}

func (p *parser) advance() {
	if p.index < p.end {
		p.index++
	}
}

func (p *parser) at(index int) html_lexer.Token {
	if index < p.end {
		return p.tokens[index]
	}
	if p.end < len(p.tokens) {
		return html_lexer.Token{
			Kind:  html_lexer.TEndOfFile,
			Range: logger.Range{Loc: p.tokens[p.end].Range.Loc},
		}
	}
	return html_lexer.Token{
		Kind:  html_lexer.TEndOfFile,
		Range: logger.Range{Loc: logger.Loc{Start: int32(len(p.source.Contents))}},
	}
}

func (p *parser) current() html_lexer.Token {
	return p.at(p.index)
}

func (p *parser) next() html_lexer.Token {
	return p.at(p.index + 1)
}

func (p *parser) raw() string {
	t := p.current()
	return p.source.Contents[t.Range.Loc.Start:t.Range.End()]
}

func (p *parser) peek(kind html_lexer.T) bool {
	return kind == p.current().Kind
}

func (p *parser) eat(kind html_lexer.T) bool {
	if p.peek(kind) {
		p.advance()
		return true
	}
	return false
}

func (p *parser) expect(kind html_lexer.T) bool {
	if p.eat(kind) {
		return true
	}
	t := p.current()
	var text string

	switch t.Kind {
	case html_lexer.TEndOfFile:
		text = fmt.Sprintf("Expected %s but found %s", kind.String(), t.Kind.String())
		t.Range.Len = 0
	default:
		text = fmt.Sprintf("Expected %s but found %q", kind.String(), p.raw())
	}

	if t.Range.Loc.Start > p.prevError.Start {
		p.log.AddRangeWarning(&p.source, t.Range, text)
		p.prevError = t.Range.Loc
	}
	return false
}

func (p *parser) unexpected() {
	if t := p.current(); t.Range.Loc.Start > p.prevError.Start {
		var text string
		switch t.Kind {
		case html_lexer.TEndOfFile:
			text = fmt.Sprintf("Unexpected %s", t.Kind.String())
			t.Range.Len = 0
		default:
			text = fmt.Sprintf("Unexpected %s", p.raw())
		}
		p.log.AddRangeWarning(&p.source, t.Range, text)
		p.prevError = t.Range.Loc
	}
}

func (p *parser) parseDoctype() html_ast.Doctype {

	p.expect(html_lexer.TDoctypeStart)

	t := p.current()

	if t.Kind == html_lexer.TText {
		name := p.decoded()
		p.advance()

	}

	p.expect(html_lexer.TText)
	p.expect(html_lexer.TTagEnd)

	return html_ast.Doctype{
		Name:  "name",
		Value: "value",
	}
}
