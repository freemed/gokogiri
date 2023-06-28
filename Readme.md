Gokogiri
========
[![Build Status](https://travis-ci.org/freemed/gokogiri.svg?branch=master)](https://travis-ci.org/freemed/gokogiri)
[![codecov](https://codecov.io/gh/freemed/gokogiri/branch/master/graph/badge.svg)](https://codecov.io/gh/freemed/gokogiri)
[![Go Report Card](https://goreportcard.com/badge/github.com/freemed/gokogiri)](https://goreportcard.com/report/github.com/freemed/gokogiri)
[![GoDoc](https://godoc.org/github.com/freemed/gokogiri?status.svg)](https://godoc.org/github.com/freemed/gokogiri)

LibXML bindings for the Go programming language.
------------------------------------------------
The gokogiri package provides a Go interface to the libxml2 library.

It is inspired by the ruby-based Nokogiri API, and allows one to parse, manipulate, and create HTML and XML documents. Nodes can be selected using either CSS selectors (in much the same fashion as jQuery) or XPath 1.0 expressions, and a simple DOM-like interface allows for building up documents from scratch.

It uses parsing default options that ignore errors or warnings, making it suitable for the poorly-formed 'tag soup' often found on the web. The xml.StrictParsingOption is conveniently provided for standards-compliant behaviour.

This fork incorporates changes required to compile on Go 1.4 and above.

## Installation

```bash
# Linux
sudo apt-get install libxml2-dev
# Mac
brew install libxml2

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
  "io/ioutil"
  "github.com/freemed/gokogiri"
)

func main() {
  // fetch and read a web page
  resp, _ := http.Get("http://www.google.com")
  page, _ := ioutil.ReadAll(resp.Body)

  // parse the web page
  doc, _ := gokogiri.ParseHtml(page)

  // perform operations on the parsed page -- consult the tests for examples

  // important -- don't forget to free the resources when you're done!
  doc.Free()
}
```
