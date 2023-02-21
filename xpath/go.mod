module github.com/freemed/gokogiri/xpath

go 1.18

replace (
	github.com/freemed/gokogiri/help => ../help
	github.com/freemed/gokogiri/util => ../util
)

require (
	github.com/freemed/gokogiri/help v0.0.0-20201230192900-c04779a870c8
	github.com/freemed/gokogiri/util v0.0.0-20220627154600-2acb041aa5ac
)
