// Локальный пакет flag — омоним stdlib, но с другим путём импорта.
package flag

func String(name string) *string { return &name }

func Parse() {}
