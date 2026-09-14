// Command veval reads an evaluation report and says whether it holds
// together, recomputes what it derives, and writes it in the form another
// reader wants: a document for a person, SARIF for a tool.
//
// Every subcommand answers the same three questions the same way: 0 the
// report is sound, 1 it breaks a rule and the broken rules are on standard
// output, 2 the command could not run and why is on standard error.
package main

import "os"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
