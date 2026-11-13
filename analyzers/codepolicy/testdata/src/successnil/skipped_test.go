package successnil

func skippedTestReturn() (*table, error) { return nil, nil }

func skippedTest(c client) {
	v, err := c.TableWithColumns("")
	if err != nil {
		return
	}
	if v == nil {
		return
	}
}
