// nolint
// Package pkg2 exitcheck testdata package.
package pkg2

import "os"

func Fun2() {
	os.Exit(2)
}

func main() {
	os.Exit(2)
}
