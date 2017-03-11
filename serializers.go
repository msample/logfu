package logfu

import (
	"io"

	"github.com/go-kit/kit/log"
)

// SerializerFunc type allows you convert a serializer function to a
// Serializer interface
type SerializerFunc func(io.Writer, []interface{}) error

func (o SerializerFunc) Serialize(w io.Writer, keyvals []interface{}) error {
	return o(w, keyvals)
}

// JSONSerializerFac is a factory for JSON Serializers
func JSONSerializerFac() (Serializer, error) {
	return SerializerFunc(JSONSerialize), nil
}

// JSONSerialize is the JSON serializer implementation
func JSONSerialize(w io.Writer, kvs []interface{}) error {
	return log.NewJSONLogger(w).Log(kvs...)
}

// LogfmtSerializerFac is a factory for JSON Serializers
func LogfmtSerializerFac() (Serializer, error) {
	return SerializerFunc(LogfmtSerialize), nil
}

// LogfmtSerialize is the log fmt serializer implementation
func LogfmtSerialize(w io.Writer, kvs []interface{}) error {
	return log.NewLogfmtLogger(w).Log(kvs...)
}
