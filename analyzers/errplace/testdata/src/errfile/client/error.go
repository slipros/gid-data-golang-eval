package client

type ConnectionError struct{}

func (e *ConnectionError) Error() string { return "connection failure" }
