package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func abbreviatePath(path string, maxLength int) string {
	if len(path) <= maxLength {
		return path
	}
	return path[:maxLength-3] + "..."
}

func prompt(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nFailed to read input: %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(input)
}
