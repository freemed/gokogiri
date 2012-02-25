package test

import (
	"gokogiri"
	"gokogiri/help"
	"testing"
	"strings"
)

func TestHtmlSimpleParse(t *testing.T) {
	doc := gokogiri.HtmlParseString("<html><head /><body /></html>")
	if doc.Size() != 1 {
		t.Error("Incorrect size")
	}
	// Doctype gets returned as the first child!
	htmlTag := doc.First().Next()
	if htmlTag.Size() != 2 {
		print(htmlTag.Name())
		t.Error("Two tags are inside of <html>")
	}
	doc.Free()
	help.XmlCleanUpParser()
	if help.XmlMemoryAllocation() != 0 {
		t.Errorf("Memeory leaks %d!!!", help.XmlMemoryAllocation())
		help.XmlMemoryLeakReport()
	}
}

func TestHtmlCDataTag(t *testing.T) {
	doc := gokogiri.HtmlParseString(LoadFile("docs/script.html"))
	if doc.Size() != 1 {
		t.Error("Incorrect size")
	}
	scriptTag := doc.RootElement().FirstElement().FirstElement()
	if scriptTag.Name() != "script" {
		t.Error("Should have selected the script tag")
	}
	content := scriptTag.Content()
	scriptTag.SetContent(content)
	doc.Free()
	help.XmlCleanUpParser()
	if help.XmlMemoryAllocation() != 0 {
		t.Errorf("Memeory leaks %d!!!", help.XmlMemoryAllocation())
		help.XmlMemoryLeakReport()
	}
}

func TestHtmlEmptyDoc(t *testing.T) {
	doc := gokogiri.HtmlParseString("")
	if !strings.Contains(doc.DumpHTML(), "<!DOCTYPE") {
		t.Error("Should have actually made a doc")
	}
	doc.Free()
	help.XmlCleanUpParser()
	if help.XmlMemoryAllocation() != 0 {
		t.Errorf("Memeory leaks %d!!!", help.XmlMemoryAllocation())
		help.XmlMemoryLeakReport()
	}
}

func TestSetHtmlContent(t *testing.T) {
	doc := gokogiri.HtmlParseString("<html><head /><body /></html>")
    if doc.Size() != 1 {
        t.Error("Incorrect size")
    }

	head := doc.RootElement().FirstElement()
    body := head.NextElement()
	Equal(t, head.Content(), "")
	head.SetHtmlContent("<meta class=\"beauty\">")
	Equal(t, head.Content(), "<meta class=\"beauty\">")

	Equal(t, body.Content(), "")
	body.SetHtmlContent("<script src=\"somefunnyplace.com/fun.js\">")
	Equal(t, body.Content(), "<script src=\"somefunnyplace.com/fun.js\"></script>")
	
    doc.Free()
    help.XmlCleanUpParser()
    if help.XmlMemoryAllocation() != 0 {
        t.Errorf("Memeory leaks %d!!!", help.XmlMemoryAllocation())
        help.XmlMemoryLeakReport()
    }
}

func TestAppendHtmlContent(t *testing.T) {
  doc := gokogiri.XmlParseString("<root><parent><child /></parent></root>")
  root := doc.RootElement()
  parent := root.FirstElement()
  Equal(t, parent.Size(), 1)
  parent.AppendHtmlContent(" and <sibling/>")
  Equal(t, parent.Size(), 2)
  Equal(t, parent.First().Name(), "child")
  Equal(t, parent.First().Next().Name(), "text")
  Equal(t, parent.First().Next().Content(), " and ")
  doc.Free()
    help.XmlCleanUpParser()
    if help.XmlMemoryAllocation() != 0 {
        t.Errorf("Memeory leaks %d!!!", help.XmlMemoryAllocation())
        help.XmlMemoryLeakReport()
    }
}

func TestSetMetaEncoding(t *testing.T) {
	doc := gokogiri.HtmlParseString("<html><head /><body /></html>")
	if doc.Size() != 1 {
		t.Error("Incorrect size")
	}
	doc.SetMetaEncoding("utf-8")
	// Doctype gets returned as the first child!
	htmlTag := doc.First().Next()
	Equal(t, htmlTag.First().First().String(), `<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>`)
	doc.Free()
	help.XmlCleanUpParser()
	if help.XmlMemoryAllocation() != 0 {
		t.Errorf("Memeory leaks %d!!!", help.XmlMemoryAllocation())
		help.XmlMemoryLeakReport()
	}
}


