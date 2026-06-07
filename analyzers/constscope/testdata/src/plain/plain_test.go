// Использование из *_test.go делает константу непереносимой,
// сами константы тестовых файлов не проверяются.
package plain

const fixtureLabel = "fixture"

func helperForTest() string { return testLabel + fixtureLabel }

var _ = helperForTest
