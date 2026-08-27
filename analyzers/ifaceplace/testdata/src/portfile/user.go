// Eval for GID-271 with settings.exclude: the only consumer of the Loader
// interface.
package portfile

// Consumer is the only user of Loader.
type Consumer struct {
	loader Loader
}
