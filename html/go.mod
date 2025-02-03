module github.com/freemed/gokogiri/html

go 1.22

toolchain go1.23.2

replace (
	github.com/freemed/gokogiri/help => ../help
	github.com/freemed/gokogiri/util => ../util
	github.com/freemed/gokogiri/xml => ../xml
)

require (
	github.com/freemed/gokogiri/help v0.0.0-20230628164547-0f93de0487ac
	github.com/freemed/gokogiri/util v0.0.0-20230628164547-0f93de0487ac
	github.com/freemed/gokogiri/xml v0.0.0-20230628164547-0f93de0487ac
)

require github.com/freemed/gokogiri/xpath v0.0.0-20230628164547-0f93de0487ac // indirect
