// Eval for GID-130 — order violations.
package constvarorder

import "time"

var Late = time.Second // declared before const — not a violation by itself...

const AfterVar = 1 // want `GID-130: a const block must be at the top of the file, right after import and above var, types and functions\. Fix: move it up`

type Svc struct{}

const AfterType = 2 // want `GID-130: a const block must be at the top of the file, right after import and above var, types and functions\. Fix: move it up`

func Do() {}

var AfterFunc = 3 // want `GID-130: a var block must be at the top of the file, after const and above types and functions\. Fix: move it up`
