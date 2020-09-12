//go:generate go run -tags generate gen.go

package main

import (
	"chrome"
	"log"
	"net/http"
	"sync"
)

// Go types that are bound to the App must be thread-safe, because each binding
// is executed in its own goroutine. In this simple case we may use atomic
// operations, but for more complex cases one should use proper synchronization.
type counter struct {
	sync.Mutex
	count int
}

func (c *counter) Add(n int) {
	c.Lock()
	defer c.Unlock()
	c.count = c.count + n
}

func (c *counter) Value() int {
	c.Lock()
	defer c.Unlock()
	return c.count
}

func main() {
	// New app by chrome
	app, err := chrome.New(480, 320, "http://localhost:8080", "")
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	// A simple way to know when App is ready (uses body.onload event in JS)
	app.Bind("start", func() {
		log.Println("App is ready")
	})

	// Create and bind Go object to the App
	c := &counter{}
	app.Bind("counterAdd", c.Add)
	app.Bind("counterValue", c.Value)

	// Load HTML.
	// You may also use `data:text/html,<base64>` approach to load initial HTML,
	// e.g: ui.Load("data:text/html," + url.PathEscape(html))
	//app.Load("data:text/html," + url.PathEscape(html))

	// launcher server background
	go http.ListenAndServe(":8080", http.FileServer(FS))

	// Wait until App window is closed
	<-app.Done()
	log.Println("exiting...")
}
