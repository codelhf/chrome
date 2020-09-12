# ChromeApp
ChromeApp is changed from https://godoc.org/github.com/zserge/lorca

## Example

```go
app, _ := chrome.New(480, 320, "", "")
defer app.Close()

// Bind Go function to be available in JS. Go function may be long-running and
// blocking - in JS it's represented with a Promise.
app.Bind("add", func(a, b int) int { return a + b })

// Call JS function from Go. Functions may be asynchronous, i.e. return promises
n := app.Eval(`Math.random()`).Float()
fmt.Println(n)

// Call JS that calls Go and so on and so on...
m := app.Eval(`add(2, 3)`).Int()
fmt.Println(m)

// Wait until App window is closed
<-app.Done()
```

## Hello World

Here are the steps to run the hello world example.

```
cd examples/counter
go get
go run main.go
```
![counter](examples/counter/counter.gif)

## Features

* Pure Go library (no cgo) with a very simple API
* Small application size (normally 5-10MB)
* Best of both worlds - the whole power of HTML/CSS to make your UI look
	good, combined with Go performance and ease of development
* Expose Go functions/methods and call them from JavaScript
* Call arbitrary JavaScript code from Go
* Asynchronous flow between UI and main app in both languages (async/await and Goroutines)
* Supports loading web UI from the local web server or via data URL
* Supports embedding all assets into a single binary
* Supports testing your app with the UI in the headless mode
* Supports multiple app windows
* Supports packaging and branding (e.g. custom app icons). Packaging for all
	three OS can be done on a single machine using GOOS and GOARCH variables.

Also, limitations by design:

* Requires Chrome/Chromium >= 70 to be installed.
* No control over the Chrome window yet (e.g. you can't remove border, make it
	transparent, control position or size).
* No window menu (tray menus and native OS dialogs are still possible via
	3rd-party libraries)

If you want to have more control of the browser window - consider using
[webview](https://github.com/zserge/webview) library with a similar API, so
migration would be smooth.

## How it works

Under the hood ChromeApp uses [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/) to instrument on a Chrome instance. First ChromeApp tries to locate your installed Chrome, starts a remote debugging instance binding to an ephemeral port and reads from `stderr` for the actual WebSocket endpoint. Then ChromeApp opens a new client connection to the WebSocket server, and instruments Chrome by sending JSON messages of Chrome DevTools Protocol methods via WebSocket. JavaScript functions are evaluated in Chrome, while Go functions actually run in Go runtime and returned values are sent to Chrome.

## License

Code is distributed under MIT license, feel free to use it in your proprietary
projects as well.

