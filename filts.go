package logfu

type FilterFunc func([]interface{}) ([]interface{}, error)

func (o FilterFunc) Filter(keyvals []interface{}) ([]interface{}, error) {
	return o(keyvals)
}
