package main

import (
	"fmt"
	"io"

	"github.com/mickeyyaya/v-eval/internal/version"
	"github.com/mickeyyaya/v-eval/schema"
)

// runVersion says which build this is and which report schema it reads. The
// two travel on one line because neither answers the question on its own: a
// bug report quoting this line names the binary and the contract it holds
// reports to, without anyone having to ask for the second half.
//
// It takes no operands. What it answers does not depend on any file, so a
// named file is a misunderstanding worth reporting rather than a word to
// ignore.
func runVersion(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := newFlags("version", stderr)
	if _, code := parseArgs(flags, args, 0); code != exitOK {
		return code
	}
	fmt.Fprintf(stdout, "veval %s schema %s\n", version.String(), schema.Version)
	return exitOK
}
