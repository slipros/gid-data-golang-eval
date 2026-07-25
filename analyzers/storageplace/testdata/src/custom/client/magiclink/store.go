// Positive: an in-house redis wrapper added through settings.packages.
package magiclink

import "git.example.com/go-library/eredis" // want `GID-249: package "custom/client/magiclink" reaches a data store directly — driver "git.example.com/go-library/eredis" belongs to the repository layer`

type Store struct {
	conn *eredis.Connection
}
