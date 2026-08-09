//go:build windows

package engine

import (
	"syscall"
	"unsafe"
)

// Windows 控制台模式标志
const (
	wEnableLineInput = 0x0002
	wEnableEchoInput = 0x0004
	wEnableVT        = 0x0004 // 输出模式：虚拟终端处理
	// GetStdHandle 标准句柄常量：Windows 用 (DWORD)-10/-11 的位模式
	stdInputHandle  uintptr = 0xFFFFFFF6 // (DWORD)-10
	stdOutputHandle uintptr = 0xFFFFFFF5 // (DWORD)-11
)

var (
	wKernel32  = syscall.NewLazyDLL("kernel32.dll")
	wGetStdHnd = wKernel32.NewProc("GetStdHandle")
	wGetMode   = wKernel32.NewProc("GetConsoleMode")
	wSetMode   = wKernel32.NewProc("SetConsoleMode")
)

// makeRaw 关输入行缓冲/回显，并启用输出 ANSI 转义（Windows 10 1511+）。
func makeRaw() (restore func() error, err error) {
	hIn, _, _ := wGetStdHnd.Call(stdInputHandle)
	if hIn == 0 || hIn == ^uintptr(0) {
		return nil, syscall.EINVAL
	}
	var oldIn uint32
	if r, _, _ := wGetMode.Call(hIn, uintptr(unsafe.Pointer(&oldIn))); r == 0 {
		return nil, syscall.EINVAL
	}
	newIn := oldIn &^ (wEnableLineInput | wEnableEchoInput)
	if r, _, _ := wSetMode.Call(hIn, uintptr(newIn)); r == 0 {
		return nil, syscall.EINVAL
	}
	// 输出句柄启用 VT（失败不致命，只是没有颜色）
	hOut, _, _ := wGetStdHnd.Call(stdOutputHandle)
	if hOut != 0 && hOut != ^uintptr(0) {
		var oldOut uint32
		if r, _, _ := wGetMode.Call(hOut, uintptr(unsafe.Pointer(&oldOut))); r != 0 {
			_, _, _ = wSetMode.Call(hOut, uintptr(oldOut|wEnableVT))
		}
	}
	return func() error {
		r, _, _ := wSetMode.Call(hIn, uintptr(oldIn))
		if r == 0 {
			return syscall.EINVAL
		}
		return nil
	}, nil
}
