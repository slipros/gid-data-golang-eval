// A use from *_test.go makes the constant immovable;
// constants of test files themselves are not checked.
package plain

const fixtureLabel = "fixture"

func helperForTest() string { return testLabel + fixtureLabel }

var _ = helperForTest
