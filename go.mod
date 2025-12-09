module github.com/freemed/gokogiri

go 1.24

replace (
	github.com/freemed/gokogiri/help => ./help
	github.com/freemed/gokogiri/html => ./html
	github.com/freemed/gokogiri/util => ./util
	github.com/freemed/gokogiri/xml => ./xml
	github.com/freemed/gokogiri/xpath => ./xpath
	github.com/freemed/rubex => ../rubex
)

require (
	github.com/freemed/gokogiri/help v0.0.0-20251130225105-1c0457d97f4b
	github.com/freemed/gokogiri/html v0.0.0-20251130225105-1c0457d97f4b
	github.com/freemed/gokogiri/xml v0.0.0-20251130225105-1c0457d97f4b
)

require (
	github.com/freemed/gokogiri/util v0.0.0-20251130225105-1c0457d97f4b // indirect
	github.com/freemed/gokogiri/xpath v0.0.0-20251130225105-1c0457d97f4b // indirect
)
