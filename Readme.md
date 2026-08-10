Gokogiri
========
[![Build Status](https://travis-ci.org/freemed/gokogiri.svg?branch=master)](https://travis-ci.org/freemed/gokogiri)
[![codecov](https://codecov.io/gh/freemed/gokogiri/branch/master/graph/badge.svg)](https://codecov.io/gh/freemed/gokogiri)
[![Go Report Card](https://goreportcard.com/badge/github.com/freemed/gokogiri)](https://goreportcard.com/report/github.com/freemed/gokogiri)
[![GoDoc](https://godoc.org/github.com/freemed/gokogiri?status.svg)](https://godoc.org/github.com/freemed/gokogiri)

Pure-Go XML/HTML DOM and XPath library
---------------------------------------

Gokogiri is a pure-Go library providing a DOM-like interface for XML and HTML
documents, inspired by the Ruby Nokogiri API. It allows you to parse, manipulate,
and create HTML and XML documents. Nodes can be selected using CSS selectors
(in much the same fashion as jQuery) or XPath 1.0 expressions.

Version 5.1+ is a pure-Go implementation with zero C dependencies. Earlier
versions required libxml2/libxslt C libraries.

## Installation

```bash
go get github.com/freemed/gokogiri
```

## Running tests

```bash
go test github.com/freemed/gokogiri/...
```

## Basic example

```go
package main

import (
  "net/http"
  "io"
  "github.com/freemed/gokogiri"
)

func main() {
  // fetch and read a web page
  resp, _ := http.Get("http://www.google.com")
  page, _ := io.ReadAll(resp.Body)

  // parse the web page
  doc, _ := gokogiri.ParseHtml(page)

  // perform operations on the parsed page -- consult the tests for examples

  // important -- don't forget to free the resources when you're done!
  doc.Free()
}
```
