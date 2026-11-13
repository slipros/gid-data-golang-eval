package successnil

type table struct{ ID string }

type client interface {
	TableWithColumns(string) (*table, error)
	List() ([]table, error)
	Map() (map[string]table, error)
	Value() (any, error)
	Channel() (chan int, error)
	Function() (func(), error)
}

func original(c client, id string) (*table, error) {
	table, err := c.TableWithColumns(id)
	if err != nil {
		err = mapForbiddenErr(err)
		if isNotFound(err) {
			err = notFound
		}
		return nil, err
	}
	if table == nil || table.ID != id { // want "GID-298"
		return nil, unavailable
	}
	return table, nil
}

func reverseAndNonNil(c client) {
	v, err := c.TableWithColumns("")
	if nil != err {
		return
	}
	if !(nil != v) { // want "GID-298"
		return
	}
}

func collection(c client) {
	v, err := c.List()
	if err != nil {
		return
	}
	if v == nil {
		return
	}
}

func nilMap(c client) {
	v, err := c.Map()
	if err != nil {
		return
	}
	if v == nil { // want "GID-298"
		return
	}
}

func nilInterface(c client) {
	v, err := c.Value()
	if err != nil {
		return
	}
	if v != nil { // want "GID-298"
		return
	}
}

func nilChannel(c client) {
	v, err := c.Channel()
	if err != nil {
		return
	}
	if v == nil { // want "GID-298"
		return
	}
}

func nilFunction(c client) {
	v, err := c.Function()
	if err != nil {
		return
	}
	if v == nil { // want "GID-298"
		return
	}
}

func field(c client) {
	v, err := c.TableWithColumns("")
	if err != nil {
		return
	}
	if v.ID == "" {
		return
	}
}

func continuingError(c client) {
	v, err := c.TableWithColumns("")
	if err != nil {
		consume(err)
	}
	if v == nil {
		return
	}
}

func fallback(c client) {
	v, err := c.TableWithColumns("")
	if err != nil {
		v = nil
		return
	}
	if v == nil {
		return
	}
}

func reassigned(c client) {
	v, err := c.TableWithColumns("")
	if err != nil {
		return
	}
	v = nil
	if v == nil {
		return
	}
}

func wrongError(c client, other error) {
	v, err := c.TableWithColumns("")
	if other != nil {
		return
	}
	if v == nil {
		consume(err)
	}
}

func checkBeforeError(c client) {
	v, err := c.TableWithColumns("")
	if v == nil || err != nil {
		return
	}
}

func shadow(c client) {
	v, err := c.TableWithColumns("")
	if err != nil {
		return
	}
	if v := optional(); v == nil {
		return
	}
	consume(v)
}

func mutateCondition(c client) {
	v, err := c.TableWithColumns("")
	if err != nil {
		return
	}
	if mutate(&v) || v == nil {
		return
	}
}

func nestedCondition(c client, enabled bool) {
	v, err := c.TableWithColumns("")
	if err != nil {
		return
	}
	if enabled && (v == nil || v.ID == "") { // want "GID-298"
		return
	}
}

func assigned(c client) {
	var v *table
	var err error
	v, err = c.TableWithColumns("")
	if err != nil {
		return
	}
	if v == nil { // want "GID-298"
		return
	}
}

func closure(c client) {
	consume(func() {
		v, err := c.TableWithColumns("")
		if err != nil {
			return
		}
		if v == nil { // want "GID-298"
			return
		}
	})
}

var (
	notFound    error
	unavailable error
)

func mapForbiddenErr(err error) error { return err }

func isNotFound(error) bool { return false }

func consume(any) {}

func optional() *table { return nil }

func mutate(**table) bool { return false }

type tables []table

type tableAlias = *table

func namedSliceResult() (tables, error) { return nil, nil }

func namedSlice() {
	v, err := namedSliceResult()
	if err != nil {
		return
	}
	if v == nil {
		return
	}
}

func multipleResult() (*table, *table, error) { return new(table), new(table), nil }

func multiple() {
	first, second, err := multipleResult()
	if err != nil {
		return
	}
	if first == nil || second == nil { // want "GID-298" "GID-298"
		return
	}
}

func optionalField(c client, v *table) {
	_, err := c.TableWithColumns("")
	if err != nil {
		return
	}
	if v == nil {
		return
	}
}

func parenthesized(c client) {
	v, err := c.TableWithColumns("")
	if err != nil {
		return
	}
	if v == nil { // want "GID-298"
		return
	}
}

func shadowedNil(c client, nil error) {
	v, err := c.Value()
	if err != nil {
		return
	}
	if v == nil {
		return
	}
}

func checkInitializerMutates(c client) {
	v, err := c.TableWithColumns("")
	if err != nil {
		return
	}
	if mutate(&v); v == nil {
		return
	}
}

func aliasResult() (tableAlias, error) { return new(table), nil }

func alias() {
	v, err := aliasResult()
	if err != nil {
		return
	}
	if v == nil { // want "GID-298"
		return
	}
}

func genericSliceResult[T ~[]table]() (T, error) { return nil, nil }

func genericSlice[T ~[]table]() {
	v, err := genericSliceResult[T]()
	if err != nil {
		return
	}
	if v == nil {
		return
	}
}
