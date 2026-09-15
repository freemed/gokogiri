package xml

import "testing"

// fmtXML parses src and serializes its root with the given format flags.
func fmtXML(t *testing.T, src string, format SerializationOption) string {
	t.Helper()
	doc, err := Parse([]byte(src), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	root := doc.Root()
	if root == nil {
		t.Fatalf("parse %q: nil root", src)
	}
	b, size := root.SerializeWithFormat(format, nil, nil)
	return string(b[:size])
}

// TestSerializeTextEscaping pins libxml2's text-node escaping: & < > are
// entities and a carriage return becomes &#13;, because a raw CR inside XML
// content is normalized to a linefeed by every conformant parser (which would
// silently destroy an X12 CRLF segment terminator).
//
// This is the XML_SAVE_LIBXSLT serializer, which the XSLT engine opts into; the
// plain path keeps gokogiri's historical output (see
// TestSerializeDefaultPathIsUnchanged).
func TestSerializeTextEscaping(t *testing.T) {
	format := XML_SAVE_AS_XML | XML_SAVE_NO_DECL | XML_SAVE_FORMAT | XML_SAVE_LIBXSLT
	cases := []struct{ in, want string }{
		{"<a>plain</a>", "<a>plain</a>"},
		{"<a>&amp;&lt;&gt;</a>", "<a>&amp;&lt;&gt;</a>"},
		{"<a>~&#13;</a>", "<a>~&#13;</a>"},
		{"<a>&#13;&#10;</a>", "<a>&#13;\n</a>"},
		// quotes, apostrophes and tabs are literal in text content
		{"<a>it's \"q\" \t</a>", "<a>it's \"q\" \t</a>"},
	}
	for _, c := range cases {
		if got := fmtXML(t, c.in, format); got != c.want {
			t.Errorf("text escaping %q:\n got: %q\nwant: %q", c.in, got, c.want)
		}
	}
}

// TestSerializeAttributeEscaping pins libxml2's attribute escaping, including
// the tab/LF/CR character references that keep attribute values from being
// mangled by attribute-value normalization on re-parse.
func TestSerializeAttributeEscaping(t *testing.T) {
	format := XML_SAVE_AS_XML | XML_SAVE_NO_DECL | XML_SAVE_FORMAT | XML_SAVE_LIBXSLT
	cases := []struct{ in, want string }{
		{`<a v="x"/>`, `<a v="x"/>`},
		{`<a v="&amp;&lt;&gt;"/>`, `<a v="&amp;&lt;&gt;"/>`},
		{`<a v="&quot;"/>`, `<a v="&quot;"/>`},
		{`<a v="it's"/>`, `<a v="it's"/>`},
		{`<a v="&#9;"/>`, `<a v="&#9;"/>`},
		{`<a v="&#10;"/>`, `<a v="&#10;"/>`},
		{`<a v="&#13;"/>`, `<a v="&#13;"/>`},
	}
	for _, c := range cases {
		if got := fmtXML(t, c.in, format); got != c.want {
			t.Errorf("attribute escaping %q:\n got: %q\nwant: %q", c.in, got, c.want)
		}
	}
}

// TestSerializeIndentation pins libxml2's tree indentation rule: children of
// an element are indented only when the element has no text content. Mixed
// content is written verbatim so that no whitespace is ever injected into a
// text node (consumers that slice fixed-width fields out of element text
// depend on this).
func TestSerializeIndentation(t *testing.T) {
	format := XML_SAVE_AS_XML | XML_SAVE_NO_DECL | XML_SAVE_FORMAT | XML_SAVE_LIBXSLT
	cases := []struct{ in, want string }{
		{"<a><b>1</b><c>2</c></a>", "<a>\n  <b>1</b>\n  <c>2</c>\n</a>"},
		{"<a>x</a>", "<a>x</a>"},
		{"<a>x<b>1</b></a>", "<a>x<b>1</b></a>"},
		{"<a><b>1</b>x</a>", "<a><b>1</b>x</a>"},
		{"<a><b>1</b>x<c>2</c></a>", "<a><b>1</b>x<c>2</c></a>"},
		{"<a><b><c>x</c></b></a>", "<a>\n  <b>\n    <c>x</c>\n  </b>\n</a>"},
		{"<a><b/></a>", "<a>\n  <b/>\n</a>"},
		{"<a/>", "<a/>"},
		{"<a><!--c--></a>", "<a>\n  <!--c-->\n</a>"},
		{"<a>x<!--c--></a>", "<a>x<!--c--></a>"},
	}
	for _, c := range cases {
		if got := fmtXML(t, c.in, format); got != c.want {
			t.Errorf("indentation %q:\n got: %q\nwant: %q", c.in, got, c.want)
		}
	}
	// Formatting off: no whitespace is added anywhere.
	if got := fmtXML(t, "<a><b>1</b><c>2</c></a>", XML_SAVE_AS_XML|XML_SAVE_NO_DECL|XML_SAVE_LIBXSLT); got != "<a><b>1</b><c>2</c></a>" {
		t.Errorf("unformatted output:\n got: %q", got)
	}
}

// TestSerializeAsText pins the XSLT "text" output method: character data is
// written verbatim, with no escaping and no indentation.
func TestSerializeAsText(t *testing.T) {
	format := XML_SAVE_NO_DECL | XML_SAVE_AS_TEXT | XML_SAVE_LIBXSLT
	cases := []struct{ in, want string }{
		{"<a>~&#13;&#10;</a>", "<a>~\r\n</a>"},
		{"<a>&amp;&lt;</a>", "<a>&<</a>"},
	}
	for _, c := range cases {
		if got := fmtXML(t, c.in, format); got != c.want {
			t.Errorf("text method %q:\n got: %q\nwant: %q", c.in, got, c.want)
		}
	}
}

// TestSerializeDefaultPathIsUnchanged pins the historical (non-libxslt) output:
// the expectations here are the shape the reference fixtures in xml/tests/ have
// always had, so this test fails if a serializer change leaks into the default
// path. The same inputs serialized with XML_SAVE_LIBXSLT give the libxml2 form
// asserted by the tests above.
func TestSerializeDefaultPathIsUnchanged(t *testing.T) {
	formatted := XML_SAVE_AS_XML | XML_SAVE_NO_DECL | XML_SAVE_FORMAT
	cases := []struct{ in, want string }{
		{"<a><b>1</b><c>2</c></a>", "<a>\n  <b>\n    1\n  </b>\n  <c>\n    2\n  </c>\n</a>"},
		{"<a>x</a>", "<a>\n  x\n</a>"},
		{"<a><fun></fun></a>", "<a>\n  <fun/>\n</a>"},
	}
	for _, c := range cases {
		if got := fmtXML(t, c.in, formatted); got != c.want {
			t.Errorf("default path formatting %q:\n got: %q\nwant: %q", c.in, got, c.want)
		}
	}
	// Same expectation as the reference fixture in tests/node/set_content:
	// input <foo><bar/></foo> with root.SetContent("<fun></fun>") serializes as
	// "<foo>\n  <fun></fun>\n</foo>". Pinning it here ties the default path to
	// the fixture that a serializer change once broke.
	doc, err := Parse([]byte("<foo><bar/></foo>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	defer doc.Free()
	if err := doc.Root().SetContent("<fun></fun>"); err != nil {
		t.Fatalf("SetContent: %v", err)
	}
	b, size := doc.Root().SerializeWithFormat(XML_SAVE_AS_XML|XML_SAVE_NO_DECL|XML_SAVE_FORMAT, nil, nil)
	if got := string(b[:size]); got != "<foo>\n  <fun></fun>\n</foo>" {
		t.Errorf("default path SetContent output:\n got: %q\nwant: %q", got, "<foo>\n  <fun></fun>\n</foo>")
	}
	// No XML_SAVE_FORMAT: children are concatenated with no whitespace.
	if got := fmtXML(t, "<a><b>1</b><c>2</c></a>", XML_SAVE_AS_XML|XML_SAVE_NO_DECL); got != "<a><b>1</b><c>2</c></a>" {
		t.Errorf("default path unformatted output:\n got: %q", got)
	}
}
