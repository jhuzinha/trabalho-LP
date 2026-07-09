//go:build !windows

package main

// enableWindowsANSI é um no-op fora do Windows: Linux e macOS já
// interpretam sequências ANSI nativamente no terminal.
func enableWindowsANSI() {}
