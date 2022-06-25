module github.com/freemed/gokogiri

go 1.18

replace (
	github.com/freemed/gokogiri/help => ./help
	github.com/freemed/gokogiri/html => ./html
	github.com/freemed/gokogiri/util => ./util
	github.com/freemed/gokogiri/xml => ./xml
	github.com/freemed/gokogiri/xpath => ./xpath
	github.com/freemed/rubex => ../rubex
)

require (
	github.com/freemed/gokogiri/help v0.0.0-20201230192900-c04779a870c8
	github.com/freemed/gokogiri/html v0.0.0-20201230192900-c04779a870c8
	github.com/freemed/gokogiri/xml v0.0.0-20201230192900-c04779a870c8
)

require (
	github.com/freemed/gokogiri/util v0.0.0-20201230192900-c04779a870c8 // indirect
	github.com/freemed/gokogiri/xpath v0.0.0-20201230192900-c04779a870c8 // indirect
)
