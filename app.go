package chrome

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
)

// App interface allows talking to the HTML5 App from Go.
type App interface {
	Load(url string) error
	Bind(name string, f interface{}) error
	Eval(js string) Value
	Done() <-chan struct{}
	Close() error
}

type app struct {
	chrome *chrome
	done   chan struct{}
	tmpDir string
}

var defaultChromeArgs = []string{
	"--disable-background-networking",
	"--disable-background-timer-throttling",
	"--disable-backgrounding-occluded-windows",
	"--disable-breakpad",
	"--disable-client-side-phishing-detection",
	"--disable-default-apps",
	"--disable-dev-shm-usage",
	"--disable-infobars",
	"--disable-extensions",
	"--disable-features=site-per-process",
	"--disable-hang-monitor",
	"--disable-ipc-flooding-protection",
	"--disable-popup-blocking",
	"--disable-prompt-on-repost",
	"--disable-renderer-backgrounding",
	"--disable-sync",
	"--disable-translate",
	"--disable-windows10-custom-titlebar",
	"--metrics-recording-only",
	"--no-first-run",
	"--no-default-browser-check",
	"--safebrowsing-disable-auto-update",
	"--enable-automation",
	"--password-store=basic",
	"--use-mock-keychain",
}

// New returns a new HTML5 App for the given URL, user profile directory, window
// size and other options passed to the browser engine. If URL is an empty
// string - a blank page is displayed. If user profile directory is an empty
// string - a temporary directory is created and it will be removed on
// app.Close(). You might want to use "--headless" custom CLI argument to test
// your App code.
func New(width, height int, url, dir string, customArgs ...string) (App, error) {
	if url == "" {
		url = "data:text/html,<html></html>"
	}
	if dir == "" {
		name, err := ioutil.TempDir("", "local")
		if err != nil {
			return nil, err
		}
		dir = name
	}
	args := append(defaultChromeArgs, fmt.Sprintf("--app=%s", url))
	args = append(args, fmt.Sprintf("--user-data-dir=%s", dir))
	args = append(args, fmt.Sprintf("--window-size=%d,%d", width, height))
	args = append(args, customArgs...)
	args = append(args, "--remote-debugging-port=0")

	chrome, err := newChromeWithArgs(LocateChrome(), args...)
	done := make(chan struct{})
	if err != nil {
		return nil, err
	}

	go func() {
		chrome.cmd.Wait()
		close(done)
	}()
	return &app{chrome: chrome, done: done, tmpDir: dir}, nil
}

func (a *app) Load(url string) error {
	return a.chrome.load(url)
}

func (a *app) Bind(name string, f interface{}) error {
	v := reflect.ValueOf(f)
	// f must be a function
	if v.Kind() != reflect.Func {
		return errors.New("only functions can be bound")
	}
	// f must return either value and error or just error
	if n := v.Type().NumOut(); n > 2 {
		return errors.New("function may only return a value or a value+error")
	}

	return a.chrome.bind(name, func(raw []json.RawMessage) (interface{}, error) {
		if len(raw) != v.Type().NumIn() {
			return nil, errors.New("function arguments mismatch")
		}
		args := []reflect.Value{}
		for i := range raw {
			arg := reflect.New(v.Type().In(i))
			if err := json.Unmarshal(raw[i], arg.Interface()); err != nil {
				return nil, err
			}
			args = append(args, arg.Elem())
		}
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		res := v.Call(args)
		switch len(res) {
		case 0:
			// No results from the function, just return nil
			return nil, nil
		case 1:
			// One result may be a value, or an error
			if res[0].Type().Implements(errorType) {
				if res[0].Interface() != nil {
					return nil, res[0].Interface().(error)
				}
				return nil, nil
			}
			return res[0].Interface(), nil
		case 2:
			// Two results: first one is value, second is error
			if !res[1].Type().Implements(errorType) {
				return nil, errors.New("second return value must be an error")
			}
			if res[1].Interface() == nil {
				return res[0].Interface(), nil
			}
			return res[0].Interface(), res[1].Interface().(error)
		default:
			return nil, errors.New("unexpected number of return values")
		}
	})
}

func (a *app) Eval(js string) Value {
	v, err := a.chrome.eval(js)
	return value{err: err, raw: v}
}

func (a *app) Done() <-chan struct{} {
	return a.done
}

func (a *app) Close() error {
	// ignore err, as the chrome process might be already dead, when user close the window.
	a.chrome.kill()
	<-a.done
	if a.tmpDir != "" {
		if err := os.RemoveAll(a.tmpDir); err != nil {
			return err
		}
	}
	return nil
}
