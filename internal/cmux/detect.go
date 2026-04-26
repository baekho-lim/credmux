package cmux

import (
	"errors"
	"os"
)

var ErrNotDetected = errors.New("cmux not detected")

func SocketPath() string {
	if p := os.Getenv("CMUX_SOCKET_PATH"); p != "" {
		return p
	}
	return "/tmp/cmux.sock"
}

func Detect() bool {
	info, err := os.Stat(SocketPath())
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSocket != 0
}
