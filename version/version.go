package version

import (
	_ "embed"
	"fmt"
)

//go:embed VERSION
var VERSION string

// PrintVersion prints the version of the application.
func PrintVersion() {
	fmt.Printf("spoofdpi v%s\n", VERSION)
	fmt.Println("A simple and fast anti-censorship tool written in Go.")
	fmt.Println("https://github.com/bariiss/SpoofDPI")
}
