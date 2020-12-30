module github.com/freemed/gokogiri

go 1.15

replace (
	github.com/freemed/gokogiri/help => ./help
	github.com/freemed/gokogiri/html => ./html
	github.com/freemed/gokogiri/util => ./util
	github.com/freemed/gokogiri/xml => ./xml
	github.com/freemed/gokogiri/xpath => ./xpath
)

require (
	github.com/freemed/gokogiri/help v0.0.0-00010101000000-000000000000
	github.com/freemed/gokogiri/html v0.0.0-00010101000000-000000000000
	github.com/freemed/gokogiri/xml v0.0.0-00010101000000-000000000000
	github.com/freemed/gokogiri/xpath v0.0.0-00010101000000-000000000000 // indirect
)
