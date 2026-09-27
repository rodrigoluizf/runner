//go:build !windows

package simulador

import "syscall"

// isAlive verifica se o processo com o PID indicado está em execução.
func isAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
