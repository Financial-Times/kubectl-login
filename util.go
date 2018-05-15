// +build !darwin

package main

import (
	"bufio"
	"os"
	"strings"
)

func readTokens() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}
