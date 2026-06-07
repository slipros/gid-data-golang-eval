// Класс «позитив»: flag.Parse (и flag.FlagSet) в библиотеке запрещены.
package libparse

import "flag"

func init() {
	flag.Parse() // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`
}

// Метод *flag.FlagSet вне main тоже запрещён.
func custom() {
	fs := flag.NewFlagSet("svc", flag.ContinueOnError) // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`
	fs.String("addr", "", "addr")                      // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`
}
