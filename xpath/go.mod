module github.com/freemed/gokogiri/xpath

go 1.16

replace (
	github.com/freemed/gokogiri => ../
	github.com/freemed/gokogiri/help => ../help
	github.com/freemed/gokogiri/util => ../util
	github.com/freemed/gokogiri/xml => ../xml
)

require (
	github.com/freemed/gokogiri/help v0.0.0-00010101000000-000000000000
	github.com/freemed/gokogiri/util v0.0.0-00010101000000-000000000000
)
