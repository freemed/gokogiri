module github.com/freemed/gokogiri/html

go 1.18

replace (
	github.com/freemed/gokogiri/help => ../help
	github.com/freemed/gokogiri/util => ../util
	github.com/freemed/gokogiri/xml => ../xml
)

require (
	github.com/freemed/gokogiri/help v0.0.0-20220627154600-2acb041aa5ac
	github.com/freemed/gokogiri/util v0.0.0-20220627154600-2acb041aa5ac
	github.com/freemed/gokogiri/xml v0.0.0-20220627154600-2acb041aa5ac
)

require github.com/freemed/gokogiri/xpath v0.0.0-20220627154600-2acb041aa5ac // indirect
