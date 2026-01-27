module github.com/freemed/gokogiri/html

go 1.24

toolchain go1.24.3

replace (
	github.com/freemed/gokogiri/help => ../help
	github.com/freemed/gokogiri/util => ../util
	github.com/freemed/gokogiri/xml => ../xml
)

require (
	github.com/freemed/gokogiri/help v0.0.0-20251209120151-edc422feefb4
	github.com/freemed/gokogiri/util v0.0.0-20251209120151-edc422feefb4
	github.com/freemed/gokogiri/xml v0.0.0-20251209120151-edc422feefb4
)

require github.com/freemed/gokogiri/xpath v0.0.0-20251209120151-edc422feefb4 // indirect
