package html_ast

import (
	"github.com/evanw/esbuild/internal/ast"
	"github.com/evanw/esbuild/internal/logger"
)

type AST struct {
	ImportRecords []ast.ImportRecord
	Doctype       Doctype
}

type Node struct {
	Loc  logger.Loc
	Data N
}

// This interface is never called. Its purpose is to encode a variant type in
// Go's type system.
type N interface {
	isNode()
}

type Doctype struct {
	Name  string
	Value string // Optional, may be ""
}

type NComment struct {
	Text string
}

type NText struct {
	Value string
}

type NElement struct {
	TagName    string
	Properties []Property
	Children   []Node
}

type Property struct {
	Key   string
	Value string
}

func (*NComment) isNode() {}
func (*NText) isNode()    {}
func (*NElement) isNode() {}
