package html_lexer

import (
	"testing"

	"github.com/evanw/esbuild/internal/logger"
	"github.com/evanw/esbuild/internal/test"
)

func lexToken(contents string) (T, string) {
	log := logger.NewDeferLog()
	tokens := Tokenize(log, test.SourceForTest(contents))
	if len(tokens) > 0 {
		t := tokens[0]
		return t.Kind, contents[t.Range.Loc.Start:t.Range.End()]
	}
	return TEndOfFile, ""
}

func lexerError(contents string) string {
	log := logger.NewDeferLog()
	Tokenize(log, test.SourceForTest(contents))
	text := ""
	for _, msg := range log.Done() {
		text += msg.String(logger.OutputOptions{}, logger.TerminalInfo{})
	}
	return text
}

func TestTokens(t *testing.T) {
	expected := []struct {
		contents string
		token    T
	}{
		{"", TEndOfFile},
		{"<!doctype", TDoctypeStart},
		{"<!--", TCommentStart},
		{"</", TTagCloseStart},
		{"<tag", TTagOpenStart},
		{"-->", TCommentEnd},
		{"/>", TTagSelfClosingEnd},
		{">", TTagEnd},
		{"-", TText},
		{"--", TText},
		{"<!", TText},
		{"!", TText},
		{"/", TText},
		{"=", TEquals},
		{"name", TName},
		{"namepace:name", TName},
		{"'string'", TString},
		{"\"string\"", TString},
	}

	for _, it := range expected {
		contents := it.contents
		token := it.token
		t.Run(contents, func(t *testing.T) {
			kind, _ := lexToken(contents)
			test.AssertEqual(t, kind, token)
		})
	}
}
