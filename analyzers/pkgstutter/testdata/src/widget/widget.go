// Eval for GID-193 (no-pkg-stutter): positive, negative, and boundary cases
// in the widget package.
package widget

// --- Positive cases (the symbol repeats the package name) ---

// WidgetOptions — a stuttering type: from outside it is widget.WidgetOptions.
type WidgetOptions struct { // want `GID-193: WidgetOptions repeats the package name widget\. Fix: from outside it is widget\.Options; drop the prefix`
	Size int
}

// WidgetCount — a stuttering function.
func WidgetCount() int { // want `GID-193: WidgetCount repeats the package name widget\. Fix: from outside it is widget\.Count; drop the prefix`
	return 0
}

// WidgetDefault — a stuttering variable.
var WidgetDefault = WidgetOptions{} // want `GID-193: WidgetDefault repeats the package name widget\. Fix: from outside it is widget\.Default; drop the prefix`

// WidgetMax — a stuttering constant.
const WidgetMax = 100 // want `GID-193: WidgetMax repeats the package name widget\. Fix: from outside it is widget\.Max; drop the prefix`

// --- Negative cases (clean code passes) ---

// Options — without the package prefix.
type Options struct {
	Size int
}

// Count — without the package prefix.
func Count() int {
	return 0
}

// --- Boundary cases ---

// NewWidget — a constructor, excluded in favor of GID-104.
func NewWidget() *Widget {
	return &Widget{}
}

// Widget — an exact match of the package name with no next word: no stutter.
type Widget struct {
	id int
}

// WidgetID — a method (has a receiver): reads as value.WidgetID, not matched.
func (w *Widget) WidgetID() int {
	return w.id
}

// widgetCache — an unexported symbol, not visible from outside: not matched.
var widgetCache = map[int]*Widget{}
