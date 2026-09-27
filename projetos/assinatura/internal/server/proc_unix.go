//go:build !windows

package server

import "syscall"

// daemonAttr cria um novo grupo de sessão para o processo filho,
// impedindo que o SIGHUP do terminal alcance o servidor.
func daemonAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
