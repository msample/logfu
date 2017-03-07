package logfu

import (
	"io"

	"github.com/go-kit/kit/log"
)

type SerializerFunc func(io.Writer, []interface{}) error

func (o SerializerFunc) Serialize(w io.Writer, keyvals []interface{}) error {
	return o(w, keyvals)
}

func JSONSerializerFac() Serializer {
	return SerializerFunc(JSONSerialize)
}

func JSONSerialize(w io.Writer, kvs []interface{}) error {
	return log.NewJSONLogger(w).Log(kvs...)
}

func LogfmtSerializerFac() Serializer {
	return SerializerFunc(LogfmtSerialize)
}

func LogfmtSerialize(w io.Writer, kvs []interface{}) error {
	return log.NewLogfmtLogger(w).Log(kvs...)
}
