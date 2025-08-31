module github.com/freemed/gokogiri

go 1.22

toolchain go1.23.2

replace (
	github.com/freemed/gokogiri/help => ./help
	github.com/freemed/gokogiri/html => ./html
	github.com/freemed/gokogiri/util => ./util
	github.com/freemed/gokogiri/xml => ./xml
	github.com/freemed/gokogiri/xpath => ./xpath
	github.com/freemed/rubex => ../rubex
)

require (
	github.com/freemed/gokogiri/help v0.0.0-20250402180648-1e651eb8ffcd
	github.com/freemed/gokogiri/html v0.0.0-20250402180648-1e651eb8ffcd
	github.com/freemed/gokogiri/xml v0.0.0-20250402180648-1e651eb8ffcd
)

require (
	github.com/freemed/gokogiri/util v0.0.0-20250402180648-1e651eb8ffcd // indirect
	github.com/freemed/gokogiri/xpath v0.0.0-20250402180648-1e651eb8ffcd // indirect
)
