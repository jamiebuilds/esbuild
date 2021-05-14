package html_printer

import (
	"strings"

	"github.com/evanw/esbuild/internal/ast"
	"github.com/evanw/esbuild/internal/html_ast"
)

type printer struct {
	importRecords []ast.ImportRecord
	sb            strings.Builder
}

func Print(tree html_ast.AST) string {
	p := printer{
		importRecords: tree.ImportRecords,
	}
	p.printDoctype(tree.Doctype)
	return p.sb.String()
}

func (p *printer) print(text string) {
	p.sb.WriteString(text)
}

func (p *printer) printDoctype(doctype html_ast.Doctype) {
	p.print("<!doctype ")
	p.print(doctype.Name)
	p.print(" ")
	p.print(doctype.Value)
	p.print(">")
}
