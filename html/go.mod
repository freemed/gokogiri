module github.com/freemed/gokogiri/html

go 1.15

replace (
	github.com/freemed/gokogiri => ../
	github.com/freemed/gokogiri/xml => ../xml
)

require (
	github.com/freemed/gokogiri v0.0.0-2019091701
	github.com/freemed/gokogiri/xml v0.0.0-2019091701
)
