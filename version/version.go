package version

import _ "embed"

//go:embed VERSION
var VERSION string

// PrintVersion prints the version of the application.
func PrintVersion() {
	println("spoofdpi", "v"+VERSION)
	println("A simple and fast anti-censorship tool written in Go.")
	println("https://github.com/bariiss/SpoofDPI")
}
