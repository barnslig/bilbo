package main

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/barnslig/bilbo/goldmark/wikitag"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

func testResolveHelper(t *testing.T, content string, expected string, pageExists bool) {
	r := wikiTagResolver{
		pageExists: func(path string) bool {
			return pageExists
		},
	}

	n := wikitag.Node{
		Content: []byte(content),
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	err := r.Resolve(w, &n)
	if err != nil {
		t.Fatal(err)
	}

	w.Flush()

	actual := buf.String()
	if actual != expected {
		t.Fatalf("Expected \"%v\", got \"%v\"", expected, actual)
	}
}

func TestResolve(t *testing.T) {
	testResolveHelper(t, "foo", "<a href=\"/foo\" class=\"internal absent\">foo</a>", false)
	testResolveHelper(t, "foo|bar", "<a href=\"/bar\" class=\"internal absent\">foo</a>", false)
	testResolveHelper(t, "foo", "<a href=\"/foo\" class=\"internal\">foo</a>", true)
	testResolveHelper(t, "foo|bar", "<a href=\"/bar\" class=\"internal\">foo</a>", true)
	testResolveHelper(t, "https://example.com", "<a href=\"https://example.com\" class=\"external\">https://example.com</a>", false)
	testResolveHelper(t, "foo|https://example.com", "<a href=\"https://example.com\" class=\"external\">foo</a>", true)
	testResolveHelper(t, "bar</a>|foo\"", "<a href=\"/foo%22\" class=\"internal absent\">bar&lt;/a&gt;</a>", false)
}

func TestExtractTitle(t *testing.T) {
	src := []byte("Test\n\n# Hello World\nfoobar")
	doc := goldmark.DefaultParser().Parse(text.NewReader(src))

	expected := "Hello World"

	actual, err := extractTitle(doc, src)
	if err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("Expected \"%v\", got \"%v\"", expected, actual)
	}
}
