//go:build linux

package engine

import (
	"syscall"
	"unsafe"
)

// makeRaw 把 stdin 切换为原始模式（关行缓冲/回显/信号），返回恢复函数。
// 适用于 Linux 与 Termux(Android)。
func makeRaw() (restore func() error, err error) {
	fd := syscall.Stdin
	var old syscall.Termios
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCGETS, uintptr(unsafe.Pointer(&old))); e != 0 {
		return nil, e
	}
	raw := old
	raw.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.ISIG
	raw.Iflag &^= syscall.IXON | syscall.ICRNL | syscall.ISTRIP
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCSETS, uintptr(unsafe.Pointer(&raw))); e != 0 {
		return nil, e
	}
	return func() error {
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCSETS, uintptr(unsafe.Pointer(&old)))
		return e
	}, nil
}
