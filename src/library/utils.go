package library

import (
	"fmt"
	"os/exec"
	"runtime"
)

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func DepsCheck() error {
    if runtime.GOOS != "linux" {
        return fmt.Errorf("unsupported OS detected")
    }

    deps := [...]string{
        "ss",
        "grep",
        "wc",
        "./frps",
    }

    for _, d := range deps {
        if ! commandExists(d) {
            return fmt.Errorf("%s could not be found", d)
        }
    }
    return nil
}

func CStr2Str(input []byte) string {
	return string(input[:clen(input)])
}

func clen(n []byte) int {
    for i := 0; i < len(n); i++ {
        if n[i] == 0 {
            return i
        }
    }
    return len(n)
}