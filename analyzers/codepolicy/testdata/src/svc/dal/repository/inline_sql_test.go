package repository

func testInlineSQL() {
	conn := connection{}
	conn.ScanRow(nil, nil, "SELECT 1", nil)
}
