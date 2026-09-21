package request

type Update struct{}

func (*Update) Validate() error { // want `GID-278:.*Update`
	return nil
}
