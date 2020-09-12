package main

import (
	"chrome"
	"log"
	"net/url"
)

func main() {
	// Create App with basic HTML passed via data URI
	ui, err := chrome.New(480, 320, "data:text/html,"+url.PathEscape(`
	<html>
		<head><title>Hello</title></head>
		<body><h1>Hello, world!</h1></body>
	</html>
	`), "")
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()
	// Wait until App window is closed
	<-ui.Done()
}
