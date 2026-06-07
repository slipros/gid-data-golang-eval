// Eval для GID-193 (no-pkg-stutter): позитивные, негативные и граничные кейсы
// в пакете widget.
package widget

// --- Позитивные кейсы (символ повторяет имя пакета) ---

// WidgetOptions — тип с заиканием: снаружи widget.WidgetOptions.
type WidgetOptions struct { // want `GID-193: WidgetOptions repeats the package name widget\. Fix: from outside it is widget\.Options; drop the prefix`
	Size int
}

// WidgetCount — функция с заиканием.
func WidgetCount() int { // want `GID-193: WidgetCount repeats the package name widget\. Fix: from outside it is widget\.Count; drop the prefix`
	return 0
}

// WidgetDefault — переменная с заиканием.
var WidgetDefault = WidgetOptions{} // want `GID-193: WidgetDefault repeats the package name widget\. Fix: from outside it is widget\.Default; drop the prefix`

// WidgetMax — константа с заиканием.
const WidgetMax = 100 // want `GID-193: WidgetMax repeats the package name widget\. Fix: from outside it is widget\.Max; drop the prefix`

// --- Негативные кейсы (чистый код проходит) ---

// Options — без префикса пакета.
type Options struct {
	Size int
}

// Count — без префикса пакета.
func Count() int {
	return 0
}

// --- Граничные кейсы ---

// NewWidget — конструктор, исключён в пользу GID-104.
func NewWidget() *Widget {
	return &Widget{}
}

// Widget — точное совпадение с именем пакета без следующего слова: нет заикания.
type Widget struct {
	id int
}

// WidgetID — метод (есть ресивер): читается как value.WidgetID, не матчится.
func (w *Widget) WidgetID() int {
	return w.id
}

// widgetCache — неэкспортируемый символ, снаружи не виден: не матчится.
var widgetCache = map[int]*Widget{}
