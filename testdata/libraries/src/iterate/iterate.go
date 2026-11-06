package iterate

func bad(items []struct{ ID int }) {
	for _, item := range items { // want "fork.example/helper"
		_ = item.ID
	}
}
