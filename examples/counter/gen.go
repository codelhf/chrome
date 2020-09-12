//+build generate

package main

func main() {
	// You can also run "npm build" or webpack here, or compress assets, or
	// generate manifests, or do other preparations for your assets.
	chrome.Embed("main", "examples/counter/assets.go", "examples/counter/www")
}
