package logfu

type FilterFunc func([]interface{}) ([]interface{}, error)

func (o FilterFunc) Filter(keyvals []interface{}) ([]interface{}, error) {
	return o(keyvals)
}

func IdentityFilterFac() (Filterer, error) {
	return FilterFunc(IdentityFilter), nil
}

func IdentityFilter(keyvals []interface{}) ([]interface{}, error) {
	return keyvals, nil
}
