package main

import (
	"bufio"
	"golang.org/x/sys/unix"
	"os"
	"strings"
)

const ioctlReadTermios = unix.TIOCGETA
const ioctlWriteTermios = unix.TIOCSETA

func readTokens() string {
	// putting the terminal in noncanonical mode, as on macos, in canonical mode, the max length of
	// a line is 1024 characters, which has the effect that only tokens less than 1023 characters can be read in canonical mode.
	// Here are some useful links:
	// https://unix.stackexchange.com/questions/204815/terminal-does-not-accept-pasted-or-typed-lines-of-more-than-1024-characters
	// https://linux.die.net/man/1/stty
	stdinFd := int(os.Stdin.Fd())
	termios, err := unix.IoctlGetTermios(stdinFd, ioctlReadTermios)
	if err != nil {
		logger.Fatalf("error: cannot read token from terminal: %v", err)
	}
	defer unix.IoctlSetTermios(stdinFd, ioctlWriteTermios, termios)

	newState := *termios
	newState.Lflag &^= unix.ICANON

	if err := unix.IoctlSetTermios(stdinFd, ioctlWriteTermios, &newState); err != nil {
		logger.Fatalf("error: cannot read token from terminal: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}
