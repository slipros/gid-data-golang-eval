package service

func pageSize(n int) int {
	return n
}

type PageCounter interface { // want `GID-276: interface PageCounter is declared below func pageSize \(line 3\); interfaces open the file, right after import, const and var\. Fix: move type PageCounter interface above func pageSize`
	Count() int
}
