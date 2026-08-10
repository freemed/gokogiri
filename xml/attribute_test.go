package xml

import "testing"

func TestSetValue(t *testing.T) {
	doc, err := Parse([]byte("<foo id=\"a\" myname=\"ff\"><bar class=\"shine\"/></foo>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}
	root := doc.Root()
	attributes := root.Attributes()
	if len(attributes) != 2 || attributes["myname"].String() != "ff" {
		t.Errorf("root's attributes do not match: got %d attrs, myname=%q", len(attributes), attributes["myname"].String())
	}
	child := root.FirstChild()
	childAttributes := child.Attributes()
	if len(childAttributes) != 1 || childAttributes["class"].String() != "shine" {
		t.Error("child's attributes do not match")
	}
	attributes["myname"].SetValue("new")
	if root.Attr("myname") != "new" {
		t.Error("SetValue did not update myname")
	}
	attributes["id"].Remove()
	if root.Attr("id") != "" {
		t.Error("Remove did not remove id attribute")
	}
	doc.Free()
}

func TestSetAttribute(t *testing.T) {
	doc, err := Parse([]byte("<foo id=\"a\" myname=\"ff\"><bar class=\"shine\"/></foo>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}
	root := doc.Root()
	root.SetAttr("id", "cooler")
	root.SetAttr("id2", "hot")
	root.SetAttr("id3", "")
	if root.Attr("id") != "cooler" {
		t.Errorf("SetAttr id failed: got %q", root.Attr("id"))
	}
	if root.Attr("id3") != "" {
		t.Errorf("SetAttr id3 should be empty: got %q", root.Attr("id3"))
	}
	if root.Attribute("id3") == nil {
		t.Error("Attribute id3 should not be nil even with empty value")
	}
	doc.Free()
}

func TestSetEmptyAttribute(t *testing.T) {
	doc, err := Parse([]byte("<foo id=\"a\" myname=\"ff\"><bar class=\"shine\"/></foo>"), DefaultEncodingBytes, nil, DefaultParseOption, DefaultEncodingBytes)
	if err != nil {
		t.Error("Parsing has error:", err)
		return
	}
	root := doc.Root()
	// Setting attribute with empty name is a no-op in pure Go implementation
	root.SetAttr("", "cool")
	// Verify it didn't crash and document is intact
	if root.Attr("id") != "a" {
		t.Error("existing attribute corrupted after empty-name SetAttr")
	}
	doc.Free()
}
