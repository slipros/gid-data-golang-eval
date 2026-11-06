package client

type HTTPError struct{ Status int } // want `GID-169: error type`

func (e HTTPError) Error() string { return "http failure" } // want `GID-169: method of error type`

func (e HTTPError) Temporary() bool { return e.Status >= 500 } // want `GID-169: method of error type`
