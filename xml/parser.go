package xml

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"

	"golang.org/x/net/html/charset"
)

// lineTrackingReader wraps an io.Reader and tracks the current line number.
type lineTrackingReader struct {
	r        io.Reader
	line     int
	lastByte byte
}

func newLineTrackingReader(r io.Reader) *lineTrackingReader {
	return &lineTrackingReader{r: r, line: 1}
}

func (l *lineTrackingReader) Read(p []byte) (n int, err error) {
	n, err = l.r.Read(p)
	for i := 0; i < n; i++ {
		b := p[i]
		if l.lastByte == '\n' || (l.lastByte == '\r' && b != '\n') {
			l.line++
		}
		l.lastByte = b
	}
	return
}

// parseXML handles the full layered parsing pipeline.
func parseXML(content, inEncoding, url []byte, options ParseOption, outEncoding []byte) (doc *XmlDocument, err error) {
	inEnc := string(inEncoding)
	if inEnc == "" {
		inEnc = "utf-8"
	}
	outEnc := string(outEncoding)
	if outEnc == "" {
		outEnc = "utf-8"
	}

	// Layer 1: Encoding detection and normalization
	reader, detectedEnc := detectEncoding(content, options)
	if options&XML_PARSE_IGNORE_ENC != 0 {
		detectedEnc = string(inEncoding)
	}
	// Normalize: read all decoded bytes and strip the encoding attribute
	// from the XML declaration, since the content is now UTF-8 but the
	// declaration may still reference the original encoding (e.g. ISO-8859-1).
	// Go's encoding/xml decoder will choke on the mismatch.
	normalized, encErr := io.ReadAll(reader)
	if encErr == nil {
		normalized = stripXMLEncodingDecl(normalized)
		reader = bytes.NewReader(normalized)
	}
	_ = detectedEnc

	// Layer 2: DTD pre-scan (on original content for charset correctness)
	dtdInfo := preScanDTD(content, options)

	// Layer 3: Tree parsing
	var root *InternalNode
	if options&XML_PARSE_RECOVER != 0 {
		root, err = parseXMLRecover(reader, options)
	} else {
		root, err = parseXMLStrict(reader)
	}
	if err != nil {
		// If RECOVER, try harder
		if options&XML_PARSE_RECOVER != 0 {
			root, err = parseXMLRecover(bytes.NewReader(content), options)
		}
		if err != nil {
			return nil, ERR_FAILED_TO_PARSE_XML
		}
	}

	// Create InternalDoc
	inner := &InternalDoc{
		DocType:       XML_DOCUMENT_NODE,
		Root: root,
		DTDInfo:       dtdInfo,
		Url:           string(url),
		InEncoding:    inEnc,
		OutEncoding:   outEnc,
		UnlinkedNodes: make(map[*InternalNode]bool),
	}
	if root != nil {
		root.Doc = inner
	}

	// Layer 4: Post-processing
	if dtdInfo != nil && options&XML_PARSE_DTDATTR != 0 {
		applyDefaultAttributes(root, dtdInfo)
	}
	if options&XML_PARSE_NOBLANKS != 0 {
		StripBlankNodes(root)
	}
	if options&XML_PARSE_NOCDATA != 0 {
		mergeCDATANodes(root)
	}
	if options&XML_PARSE_NSCLEAN != 0 {
		cleanRedundantNSDeclarations(root)
	}

	// Build ID index
	inner.IDIndex = buildIDIndex(root, dtdInfo)

	doc = NewDocument(inner, len(content), []byte(inEnc), []byte(outEnc))
	return
}

// detectEncoding determines the input encoding.
func detectEncoding(content []byte, options ParseOption) (io.Reader, string) {
	if options&XML_PARSE_IGNORE_ENC != 0 {
		return bytes.NewReader(content), "utf-8"
	}
	enc, name, _ := charset.DetermineEncoding(content, "")
	return enc.NewDecoder().Reader(bytes.NewReader(content)), name
}

// preScanDTD extracts DTD declarations.
func preScanDTD(content []byte, options ParseOption) *DTDInfo {
	info := &DTDInfo{
		Entities:     make(map[string]DTDEntity),
		IDAttrs:      make(map[string]string),
		DefaultAttrs: make(map[string]map[string]string),
		Notations:    make(map[string]string),
	}

	// Simple DTD scanner: look for <!ENTITY, <!ATTLIST, <!NOTATION
	text := string(content)
	lower := strings.ToLower(text)

	// Find DOCTYPE internal subset
	dtdStart := strings.Index(lower, "<!doctype")
	if dtdStart < 0 {
		return info
	}

	// Find the internal subset [...]
	bracketStart := strings.Index(text[dtdStart:], "[")
	bracketEnd := strings.LastIndex(text, "]")
	if bracketStart < 0 || bracketEnd < 0 || bracketEnd <= dtdStart+bracketStart {
		return info
	}

	subset := text[dtdStart+bracketStart+1 : bracketEnd]

	// Parse ENTITY declarations
	parseDTDEntities(subset, info)
	// Parse ATTLIST declarations
	parseDTDAttlists(subset, info)
	// Parse NOTATION declarations
	parseDTDNotations(subset, info)

	// Fetch external DTD if DTDLOAD and not NONET
	if options&XML_PARSE_DTDLOAD != 0 && options&XML_PARSE_NONET == 0 {
		// External DTD fetching — extract SYSTEM/PUBLIC identifiers from DOCTYPE
		parseExternalDTD(text[dtdStart:], info, options)
	}

	return info
}

// parseDTDEntities extracts <!ENTITY declarations.
func parseDTDEntities(subset string, info *DTDInfo) {
	i := 0
	for {
		start := indexAfter(subset, "<!entity", i)
		if start < 0 {
			break
		}
		i = start
		end := strings.IndexByte(subset[i:], '>')
		if end < 0 {
			break
		}
		decl := strings.TrimSpace(subset[i : i+end])
		i += end + 1

		// Skip parameter entities
		if strings.HasPrefix(decl, "%") {
			continue
		}

		parts := strings.Fields(decl)
		if len(parts) < 3 {
			continue
		}
		name := parts[0]

		entity := DTDEntity{}
		// Check for NDATA
		upperDecl := strings.ToUpper(decl)
		if ndataIdx := strings.Index(upperDecl, "NDATA"); ndataIdx >= 0 {
			entity.IsNDATA = true
			afterNDATA := strings.TrimSpace(decl[ndataIdx+5:])
			if spaceIdx := strings.IndexAny(afterNDATA, " \t\r\n>"); spaceIdx >= 0 {
				entity.NDATA = afterNDATA[:spaceIdx]
			} else {
				entity.NDATA = afterNDATA
			}
		}

		// Extract SYSTEM/PUBLIC
		if sysIdx := strings.Index(upperDecl, "SYSTEM"); sysIdx >= 0 {
			entity.SystemID = extractQuoted(decl[sysIdx:])
		}
		if pubIdx := strings.Index(upperDecl, "PUBLIC"); pubIdx >= 0 {
			entity.PublicID = extractQuoted(decl[pubIdx:])
		}

		// Extract internal value
		if !entity.IsNDATA && entity.SystemID == "" {
			// Internal entity: find the quoted value
			entity.Value = extractQuoted(decl)
		}

		info.Entities[name] = entity
	}
}

// parseDTDAttlists extracts <!ATTLIST declarations.
func parseDTDAttlists(subset string, info *DTDInfo) {
	i := 0
	for {
		start := indexAfter(subset, "<!attlist", i)
		if start < 0 {
			break
		}
		i = start
		end := strings.IndexByte(subset[i:], '>')
		if end < 0 {
			break
		}
		decl := strings.TrimSpace(subset[i : i+end])
		i += end + 1

		parts := strings.Fields(decl)
		if len(parts) < 4 {
			continue
		}
		elemName := parts[0]
		attrName := parts[1]
		attrType := strings.ToUpper(parts[2])

		// Track ID attributes
		if attrType == "ID" {
			info.IDAttrs[elemName] = attrName
		}

		// Extract default value
		if len(parts) >= 4 {
			defaultVal := extractQuoted(strings.Join(parts[3:], " "))
			if defaultVal != "" {
				if info.DefaultAttrs[elemName] == nil {
					info.DefaultAttrs[elemName] = make(map[string]string)
				}
				info.DefaultAttrs[elemName][attrName] = defaultVal
			}
		}
	}
}

// parseDTDNotations extracts <!NOTATION declarations.
func parseDTDNotations(subset string, info *DTDInfo) {
	i := 0
	for {
		start := indexAfter(subset, "<!notation", i)
		if start < 0 {
			break
		}
		i = start
		end := strings.IndexByte(subset[i:], '>')
		if end < 0 {
			break
		}
		decl := strings.TrimSpace(subset[i : i+end])
		i += end + 1

		parts := strings.Fields(decl)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		upperDecl := strings.ToUpper(decl)
		if sysIdx := strings.Index(upperDecl, "SYSTEM"); sysIdx >= 0 {
			info.Notations[name] = extractQuoted(decl[sysIdx:])
		}
	}
}

// parseExternalDTD fetches and processes external DTD subsets.
func parseExternalDTD(doctypeDecl string, info *DTDInfo, options ParseOption) {
	if options&XML_PARSE_NONET != 0 {
		return
	}
	// External DTD fetching would go here in a full implementation.
	// For now, we handle internal subset only.
}

// extractQuoted extracts a double-quoted or single-quoted string.
func extractQuoted(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '"' || s[i] == '\'' {
			quote := s[i]
			for j := i + 1; j < len(s); j++ {
				if s[j] == quote {
					return s[i+1 : j]
				}
			}
			break
		}
	}
	return ""
}

// indexAfter finds a substring case-insensitively starting from offset.
func indexAfter(s, substr string, offset int) int {
	remaining := s[offset:]
	idx := strings.Index(strings.ToLower(remaining), strings.ToLower(substr))
	if idx < 0 {
		return -1
	}
	return offset + idx + len(substr)
}

// parseXMLStrict parses XML using encoding/xml.
func parseXMLStrict(r io.Reader) (*InternalNode, error) {
	lr := newLineTrackingReader(r)
	decoder := xml.NewDecoder(lr)
	decoder.Strict = true
	return buildTreeFromDecoder(decoder, lr)
}

// parseXMLRecover parses XML in lenient mode.
func parseXMLRecover(data io.Reader, options ParseOption) (*InternalNode, error) {
	// Read all and try encoding/xml
	content, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}
	lr := newLineTrackingReader(bytes.NewReader(content))
	decoder := xml.NewDecoder(lr)
	decoder.Strict = false
	return buildTreeFromDecoder(decoder, lr)
}

// buildTreeFromDecoder constructs the DOM tree from an xml.Decoder.
func buildTreeFromDecoder(decoder *xml.Decoder, lr *lineTrackingReader) (root *InternalNode, err error) {
	var current *InternalNode
	depth := 0

	for {
		tok, tokErr := decoder.Token()
		if tokErr == io.EOF {
			break
		}
		if tokErr != nil {
			if err == nil {
				err = tokErr
			}
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			elem := &InternalNode{
				Typ:   XML_ELEMENT_NODE,
				Name:  t.Name.Local,
				Line:  lr.line,
				Valid: true,
			}
			// Handle namespace
			if t.Name.Space != "" {
				elem.Ns = &InternalNs{Href: t.Name.Space}
			}
			// Handle attributes
			for _, attr := range t.Attr {
				if attr.Name.Space == "xmlns" {
					elem.DeclareNamespace(attr.Name.Local, attr.Value)
				} else if attr.Name.Local == "xmlns" {
					elem.DeclareNamespace("", attr.Value)
				} else {
					a := &InternalAttr{Name: attr.Name.Local, Value: attr.Value}
					if attr.Name.Space != "" {
						a.Ns = &InternalNs{Href: attr.Name.Space}
					}
					elem.Props = append(elem.Props, a)
				}
			}
			if current != nil {
				current.AppendChild(elem)
			} else if root == nil {
				root = elem
			}
			current = elem
			depth++

		case xml.EndElement:
			if current != nil {
				current = current.Parent
			}
			depth--

		case xml.CharData:
			text := string(t)
			// Skip whitespace-only text between elements
			if current != nil {
				textNode := &InternalNode{
					Typ:     XML_TEXT_NODE,
					Content: text,
					Line:    lr.line,
					Valid:   true,
				}
				current.AppendChild(textNode)
			} else if root == nil && strings.TrimSpace(text) != "" {
				// Top-level text
				root = &InternalNode{
					Typ:     XML_TEXT_NODE,
					Content: text,
					Line:    lr.line,
					Valid:   true,
				}
			}

		case xml.Comment:
			commentNode := &InternalNode{
				Typ:     XML_COMMENT_NODE,
				Content: string(t),
				Line:    lr.line,
				Valid:   true,
			}
			if current != nil {
				current.AppendChild(commentNode)
			}
			// Comment before root element: skip (not part of DOM tree)

		case xml.ProcInst:
			piNode := &InternalNode{
				Typ:     XML_PI_NODE,
				Name:    t.Target,
				Content: string(t.Inst),
				Line:    lr.line,
				Valid:   true,
			}
			if current != nil {
				current.AppendChild(piNode)
			}
			// XML declaration before root element: skip (not part of DOM tree)

		case xml.Directive:
			// <!DOCTYPE ...> — stored as directive, not in tree
			// CDATA sections come through as CharData with encoding/xml
			dirText := string(t)
			if strings.HasPrefix(dirText, "[CDATA[") {
				cdataContent := dirText[7 : len(dirText)-2]
				cdataNode := &InternalNode{
					Typ:     XML_CDATA_SECTION_NODE,
					Content: cdataContent,
					Line:    lr.line,
					Valid:   true,
				}
				if current != nil {
					current.AppendChild(cdataNode)
				}
			}
		}
	}

	return root, nil
}

// Layer 4 post-processing functions

// applyDefaultAttributes applies DTD default attribute values.
func applyDefaultAttributes(root *InternalNode, info *DTDInfo) {
	if root == nil || info == nil {
		return
	}
	var walk func(n *InternalNode)
	walk = func(n *InternalNode) {
		if n.Typ == XML_ELEMENT_NODE {
			if defaults, ok := info.DefaultAttrs[n.Name]; ok {
				for attrName, attrVal := range defaults {
					if _, found := n.GetAttr(attrName); !found {
						n.SetAttr(attrName, attrVal)
					}
				}
			}
		}
		for c := n.Children; c != nil; c = c.Next {
			walk(c)
		}
	}
	walk(root)
}

// stripBlankNodes removes whitespace-only text nodes.
func StripBlankNodes(root *InternalNode) {
	if root == nil {
		return
	}
	var walk func(n *InternalNode, removeList *[]*InternalNode)
	removeList := new([]*InternalNode)
	walk = func(n *InternalNode, rl *[]*InternalNode) {
		for c := n.Children; c != nil; c = c.Next {
			if c.Typ == XML_TEXT_NODE && strings.TrimSpace(c.Content) == "" {
				*rl = append(*rl, c)
			}
			walk(c, rl)
		}
	}
	walk(root, removeList)
	for _, n := range *removeList {
		n.Detach()
	}
}

// mergeCDATANodes merges CDATA nodes into adjacent text nodes.
func mergeCDATANodes(root *InternalNode) {
	if root == nil {
		return
	}
	var convertList []*InternalNode
	var walk func(n *InternalNode)
	walk = func(n *InternalNode) {
		for c := n.Children; c != nil; c = c.Next {
			if c.Typ == XML_CDATA_SECTION_NODE {
				convertList = append(convertList, c)
			}
			walk(c)
		}
	}
	walk(root)
	for _, n := range convertList {
		n.Typ = XML_TEXT_NODE
	}
}

// cleanRedundantNSDeclarations removes namespace declarations where the prefix
// is already declared with the same URI in a parent scope.
func cleanRedundantNSDeclarations(root *InternalNode) {
	if root == nil {
		return
	}
	var walk func(n *InternalNode)
	walk = func(n *InternalNode) {
		var cleaned []*InternalNs
		for _, ns := range n.NsDef {
			// Check if parent already declares this prefix with same URI
			duplicate := false
			for p := n.Parent; p != nil; p = p.Parent {
				for _, pns := range p.NsDef {
					if pns.Prefix == ns.Prefix && pns.Href == ns.Href {
						duplicate = true
						break
					}
				}
				if duplicate {
					break
				}
			}
			if !duplicate {
				cleaned = append(cleaned, ns)
			}
		}
		n.NsDef = cleaned
		for c := n.Children; c != nil; c = c.Next {
			walk(c)
		}
	}
	walk(root)
}

// buildIDIndex creates an ID attribute value → element map.
func buildIDIndex(root *InternalNode, info *DTDInfo) map[string]*InternalNode {
	if root == nil {
		return nil
	}
	index := make(map[string]*InternalNode)
	var walk func(n *InternalNode)
	walk = func(n *InternalNode) {
		if n.Typ == XML_ELEMENT_NODE {
			// Check DTD-defined ID attributes
			if info != nil {
				if idAttr, ok := info.IDAttrs[n.Name]; ok {
					if val, found := n.GetAttr(idAttr); found && val != "" {
						index[val] = n
					}
				}
			}
			// Also check for attributes literally named "id" or "ID"
			for _, a := range n.Props {
				if strings.ToLower(a.Name) == "id" && a.Ns == nil && a.Value != "" {
					if _, exists := index[a.Value]; !exists {
						index[a.Value] = n
					}
				}
			}
		}
		for c := n.Children; c != nil; c = c.Next {
			walk(c)
		}
	}
	walk(root)
	if len(index) == 0 {
		return nil
	}
	return index
}

// stripXMLEncodingDecl removes or normalizes the encoding pseudo-attribute
// from the XML declaration. After charset detection and decoding, the content
// is UTF-8, but the XML declaration may still reference the original encoding
// (e.g. ISO-8859-1). Go's encoding/xml decoder will fail with "encoding
// declared but Decoder.CharsetReader is nil" if the declared encoding
// doesn't match the actual UTF-8 byte stream.
func stripXMLEncodingDecl(data []byte) []byte {
	// Only process if it starts with XML declaration
	if len(data) < 5 || !bytes.HasPrefix(data, []byte("<?xml")) {
		return data
	}
	end := bytes.Index(data, []byte("?>"))
	if end < 0 {
		return data
	}
	decl := data[:end+2]
	rest := data[end+2:]

	// Remove encoding="..." or encoding='...' from the declaration
	var result []byte
	i := 0
	for i < len(decl) {
		// Look for encoding=
		if i+9 < len(decl) && strings.EqualFold(string(decl[i:i+9]), "encoding=") {
			quote := decl[i+9]
			// Skip encoding="..."
			if quote == '"' || quote == '\'' {
				j := i + 10
				for j < len(decl) && decl[j] != quote {
					j++
				}
				if j < len(decl) {
					j++ // skip closing quote
				}
				i = j
				continue
			}
		}
		result = append(result, decl[i])
		i++
	}
	return append(result, rest...)
}
