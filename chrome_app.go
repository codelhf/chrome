package chrome

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

// ChromeApp returns a string which points to the preferred Chrome
// executable file.
var ChromeApp = LocateChrome

// LocateChrome returns a path to the Chrome binary, or an empty string if
// Chrome installation is not found.
func LocateChrome() string {

	// If env variable "LORCACHROME" specified and it exists
	if path, ok := os.LookupEnv("LORCACHROME"); ok {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	var paths []string
	switch runtime.GOOS {
	case "windows":
		paths = []string{
			os.Getenv("LocalAppData") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("ProgramFiles") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("LocalAppData") + "/Chromium/Application/chrome.exe",
			os.Getenv("ProgramFiles") + "/Chromium/Application/chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "/Chromium/Application/chrome.exe",
		}
	case "darwin":
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
		}
	default:
		paths = []string{
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		return path
	}
	return ""
}

// PromptDownload asks user if he wants to download and install Chrome, and
// opens a download web page if the user agrees.
func PromptDownload() {
	title := "Chrome not found"
	text := "No Chrome/Chromium installation was found. Would you like to download and install it now?"

	// Ask user for confirmation
	if !messageBox(title, text) {
		return
	}

	// Open download page
	url := "https://www.google.com/chrome/"
	switch runtime.GOOS {
	case "windows":
		r := strings.NewReplacer("&", "^&")
		exec.Command("cmd", "/c", "start", r.Replace(url)).Run()
	case "darwin":
		exec.Command("open", url).Run()
	case "linux":
		exec.Command("xdg-open", url).Run()
	}
}

func messageBox(title, text string) bool {
	switch runtime.GOOS {
	case "windows":
		user32 := syscall.NewLazyDLL("user32.dll")
		messageBoxW := user32.NewProc("MessageBoxW")
		mbYesNo := 0x00000004
		mbIconQuestion := 0x00000020
		idYes := 6
		ret, _, _ := messageBoxW.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
			uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), uintptr(uint(mbYesNo|mbIconQuestion)))
		return int(ret) == idYes
	case "darwin":
		script := `set T to button returned of ` +
			`(display dialog "%s" with title "%s" buttons {"No", "Yes"} default button "Yes")`
		out, err := exec.Command("osascript", "-e", fmt.Sprintf(script, text, title)).Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				return exitError.Sys().(syscall.WaitStatus).ExitStatus() == 0
			}
		}
		return strings.TrimSpace(string(out)) == "Yes"
	case "linux":
		err := exec.Command("zenity", "--question", "--title", title, "--text", text).Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				return exitError.Sys().(syscall.WaitStatus).ExitStatus() == 0
			}
		}
		return false
	}
	return false
}
