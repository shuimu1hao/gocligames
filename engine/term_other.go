//go:build !linux && !darwin && !windows

package engine

import "errors"

// makeRaw 在其他平台不启用原始模式（游戏仍可运行，按键需回车确认）。
func makeRaw() (restore func() error, err error) {
	return func() error { return nil }, errors.New("raw mode not supported on this platform")
}
