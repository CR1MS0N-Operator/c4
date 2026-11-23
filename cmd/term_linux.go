package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

func ioctlTermios(fd uintptr, req uintptr, t *syscall.Termios) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, uintptr(unsafe.Pointer(t))); errno != 0 {
		return errno
	}
	return nil
}

// isTerminal reports whether f is a terminal.
func isTerminal(f *os.File) bool {
	var t syscall.Termios
	return ioctlTermios(f.Fd(), syscall.TCGETS, &t) == nil
}

// readHidden prints prompt to stderr and reads one line from the terminal f
// with echo disabled.
func readHidden(f *os.File, prompt string) (string, error) {
	fd := f.Fd()
	var old syscall.Termios
	if err := ioctlTermios(fd, syscall.TCGETS, &old); err != nil {
		return "", fmt.Errorf("read terminal settings: %w", err)
	}
	noEcho := old
	noEcho.Lflag &^= syscall.ECHO
	noEcho.Lflag |= syscall.ICANON
	if err := ioctlTermios(fd, syscall.TCSETS, &noEcho); err != nil {
		return "", fmt.Errorf("disable terminal echo: %w", err)
	}

	// A signal skips deferred restores; restore the terminal first, then
	// re-raise so the process still dies with the signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s, ok := <-sigCh
		if !ok {
			return
		}
		_ = ioctlTermios(fd, syscall.TCSETS, &old)
		sig := s.(syscall.Signal)
		signal.Reset(sig)
		_ = syscall.Kill(syscall.Getpid(), sig)
	}()

	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
		_ = ioctlTermios(fd, syscall.TCSETS, &old)
		fmt.Fprintln(os.Stderr)
	}()

	fmt.Fprint(os.Stderr, prompt)
	return readLine(bufio.NewReader(f))
}
