// Класс «неприменимость»: пакет без единого make(map) — анализатору нечего ловить.
package nomap

func sum(src []int) int {
	total := 0
	for _, v := range src {
		total += v
	}
	return total
}

func double(src []int) []int {
	out := make([]int, 0, len(src))
	for _, v := range src {
		out = append(out, v*2)
	}
	return out
}
