package orderpb

type CreateOrderRequest struct {
	Title  string
	Status string
}

func (*CreateOrderRequest) ProtoReflect() any { return nil }

type Order struct {
	ID     string
	Status string
	Title  string
}

func (*Order) ProtoReflect() any { return nil }

type Response struct {
	Order *Order
	Error string
}

func (*Response) ProtoReflect() any { return nil }
