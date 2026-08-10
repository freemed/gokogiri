module github.com/freemed/gokogiri

go 1.25.0

replace (
	github.com/freemed/gokogiri/help => ./help
	github.com/freemed/gokogiri/html => ./html
	github.com/freemed/gokogiri/mem => ./mem
	github.com/freemed/gokogiri/util => ./util
	github.com/freemed/gokogiri/xml => ./xml
	github.com/freemed/gokogiri/xpath => ./xpath
)

require (
	github.com/freemed/gokogiri/html v0.0.0-20260127145523-0d7d36b651ea
	github.com/freemed/gokogiri/xml v0.0.0-20260127145523-0d7d36b651ea
)

require (
	github.com/freemed/gokogiri/help v0.0.0-20260127145523-0d7d36b651ea // indirect
	github.com/freemed/gokogiri/xpath v0.0.0-20260127145523-0d7d36b651ea // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)
