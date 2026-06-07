// Класс «позитив»: flag.Parse (и flag.FlagSet) в библиотеке запрещены.
package libparse

import "flag"

func init() {
	flag.Parse() // want `GID-192: регистрация флага вне пакета main запрещена — флаги объявляет бинарь, библиотека принимает параметры`
}

// Метод *flag.FlagSet вне main тоже запрещён.
func custom() {
	fs := flag.NewFlagSet("svc", flag.ContinueOnError) // want `GID-192: регистрация флага вне пакета main запрещена — флаги объявляет бинарь, библиотека принимает параметры`
	fs.String("addr", "", "addr")                      // want `GID-192: регистрация флага вне пакета main запрещена — флаги объявляет бинарь, библиотека принимает параметры`
}
