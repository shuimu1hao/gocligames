//go:build darwin

package engine

import (
	"syscall"
	"unsafe"
)

// makeRaw 把 stdin 切换为原始模式（macOS 使用 TIOCGETA/TIOCSETA）。
func makeRaw() (restore func() error, err error) {
	fd := syscall.Stdin
	var old syscall.Termios
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGETA, uintptr(unsafe.Pointer(&old))); e != 0 {
		return nil, e
	}
	raw := old
	raw.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.ISIG
	raw.Iflag &^= syscall.IXON | syscall.ICRNL | syscall.ISTRIP
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(&raw))); e != 0 {
		return nil, e
	}
	return func() error {
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(&old)))
		return e
	}, nil
}
