// "Positive" class: flag.Parse (and flag.FlagSet) are forbidden in a library.
package libparse

import "flag"

func init() {
	flag.Parse() // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`
}

// A *flag.FlagSet method outside main is forbidden too.
func custom() {
	fs := flag.NewFlagSet("svc", flag.ContinueOnError) // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`
	fs.String("addr", "", "addr")                      // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`
}
