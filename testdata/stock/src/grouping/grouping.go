package grouping

const Topic = "work-order.events"
const Group = "advertising-api-work-order-events" // want "should only use a single global 'const' declaration"

func local() string {
	const value = "local"
	return value
}
