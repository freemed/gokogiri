module github.com/freemed/gokogiri

go 1.15

replace (
	github.com/freemed/gokogiri/html => ./html
	github.com/freemed/gokogiri/xml => ./xml
)

require (
	github.com/freemed/gokogiri/html v0.0.0-00010101000000-000000000000
	github.com/freemed/gokogiri/xml v0.0.0-00010101000000-000000000000
)
