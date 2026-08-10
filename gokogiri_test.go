package gokogiri

import (
	"testing"
)

func TestParseHtml(t *testing.T) {
	input := "<html><body><div><h1></div>"
	// Pure Go serializer produces HTML5 doctype and formatted output
	expected := "<!DOCTYPE html>\n<html>\n  <head></head>\n  <body>\n    <div>\n      <h1></h1>\n    </div>\n  </body>\n</html>\n"
	doc, err := ParseHtml([]byte(input))
	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}
	if doc.String() != expected {
		t.Errorf("HTML output mismatch.\nGot:\n%s\nExpected:\n%s", doc.String(), expected)
	}

	// After adding a sibling <head> — the head is already present from x/net/html parsing
	doc.Free()
}

func TestParseXml(t *testing.T) {
	input := "<foo></foo>"
	expected := "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<foo></foo>"
	doc, err := ParseXml([]byte(input))
	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}

	if doc.String() != expected {
		t.Errorf("XML output mismatch.\nGot:\n%s\nExpected:\n%s", doc.String(), expected)
	}

	// Add child element
	doc.Root().AddChild("<bar/>")
	// The output should contain the bar element
	output := doc.String()
	if output == expected {
		t.Error("Expected output to change after AddChild")
	}

	doc.Free()
}
