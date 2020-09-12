package chrome

import (
	"os/exec"
	"testing"
)

func TestLocate(t *testing.T) {
	exe := ChromeApp()
	if exe == "" {
		t.Fatal()
	}
	t.Log(exe)
	b, err := exec.Command(exe, "--version").CombinedOutput()
	t.Log(string(b))
	t.Log(err)
}
