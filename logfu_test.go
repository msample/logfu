package logfu_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/msample/log2"
	"github.com/msample/logfu"
)

func ExampleSetup1() {
	m := []logfu.Mode{
		{
			log2.ERROR: []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
			log2.WARN:  []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
		},
		{
			log2.ERROR: []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
			log2.WARN:  []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
			log2.DEBUG: []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
			log2.INFO:  []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
		},
	}

	lf, err := logfu.New(
		[]logfu.FiltererFac{filter1Fac},
		[]logfu.SerializerFac{serializer1Fac, serializer2Fac},
		[]logfu.WriterFac{StdoutFac, StderrFac},
		m,
		false)

	if err != nil {
		fmt.Println(err)
	}

	err = lf.ChangeToMode(0, true, true)
	if err != nil {
		fmt.Println(err)
	}
	log2.Error("msg", "err msg 1")
	log2.Warn("msg", "warn msg 1")
	log2.Info("msg", "info msg 1")
	log2.Debug("msg", "debug msg 1")
	log2.Audit("msg", "audit msg 1")

	err = lf.NextMode()
	if err != nil {
		fmt.Println(err)
	}
	log2.Error("msg", "err msg 2")
	log2.Warn("msg", "warn msg 2")
	log2.Info("msg", "info msg 2")
	log2.Debug("msg", "debug msg 2")
	log2.Audit("msg", "audit msg 2")

	err = lf.NextMode()
	if err != nil {
		fmt.Println(err)
	}
	log2.Error("msg", "err msg 3")
	log2.Warn("msg", "warn msg 3")
	log2.Info("msg", "info msg 3")
	log2.Debug("msg", "debug msg 3")
	log2.Audit("msg", "audit msg 3")

	err = lf.NextMode()
	if err != nil {
		fmt.Println(err)
	}
	log2.Error("msg", "err msg 4")
	log2.Warn("msg", "warn msg 4")
	log2.Info("msg", "info msg 4")
	log2.Debug("msg", "debug msg 4")
	log2.Audit("msg", "audit msg 4")

	// output: [msg err msg 1]
	// [msg warn msg 1]
	// [msg err msg 2]
	// [msg warn msg 2]
	// [msg info msg 2]
	// [msg debug msg 2]
	// [msg err msg 3]
	// [msg warn msg 3]
	// [msg err msg 4]
	// [msg warn msg 4]
	// [msg info msg 4]
	// [msg debug msg 4]
}

func filter1Fac() (logfu.Filterer, error) {
	return logfu.FilterFunc(filter1), nil
}

func filter1(keyvals []interface{}) ([]interface{}, error) {
	return keyvals, nil
}

func serializer1Fac() (logfu.Serializer, error) {
	return logfu.SerializerFunc(serializer1), nil
}

func serializer1(w io.Writer, keyvals []interface{}) error {
	_, err := fmt.Fprintf(w, "%v\n", keyvals)
	return err
}

func serializer2Fac() (logfu.Serializer, error) {
	return logfu.SerializerFunc(serializer2), nil
}

func serializer2(w io.Writer, keyvals []interface{}) error {
	_, err := fmt.Fprintf(w, "%#v\n", keyvals)
	return err
}

func StdoutFac() (io.Writer, error) {
	return os.Stdout, nil
	//return &Unclosable{os.Stdout}, nil
}

func StderrFac() (io.Writer, error) {
	return os.Stderr, nil
	//return &Unclosable{os.Stderr}, nil
}

type Unclosable struct {
	f io.Writer
}

func (*Unclosable) Close() error {
	return nil
}

func (o *Unclosable) Write(b []byte) (int, error) {
	return o.f.Write(b)
}

func TestNew(t *testing.T) {
	m := []logfu.Mode{
		{
			log2.ERROR: []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
			log2.WARN:  []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
		},
		{
			log2.ERROR: []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
			log2.WARN:  []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
			log2.DEBUG: []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
			log2.INFO:  []logfu.Fsw{{0, 0, 0}, {0, 1, 1}},
		},
	}

	_, err := logfu.New(
		[]logfu.FiltererFac{logfu.IdentityFilterFac},
		[]logfu.SerializerFac{logfu.JSONSerializerFac, logfu.LogfmtSerializerFac},
		[]logfu.WriterFac{logfu.StderrWriter, logfu.StdoutWriter},
		m,
		false)

	if err != nil {
		t.Errorf("expected no config failure 1: %v", err)
	}

	// introduce bad serailizer ref, check & fix
	prev := m[0][log2.ERROR][1].SerializerInd
	m[0][log2.ERROR][1].SerializerInd = 2
	_, err = logfu.New(
		[]logfu.FiltererFac{nil},
		[]logfu.SerializerFac{nil, nil},
		[]logfu.WriterFac{nil, nil},
		m,
		false)
	m[0][log2.ERROR][1].SerializerInd = prev
	if err == nil {
		t.Error("expected config failure 1")
	}

	// introduce bad filterer ref, check & fix
	prev = m[0][log2.ERROR][0].FilterInd
	m[0][log2.ERROR][0].FilterInd = -2
	_, err = logfu.New(
		[]logfu.FiltererFac{nil},
		[]logfu.SerializerFac{nil, nil},
		[]logfu.WriterFac{nil, nil},
		m,
		false)
	m[0][log2.ERROR][0].FilterInd = prev
	if err == nil {
		t.Error("expected config failure 2")
	}

	// introduce bad writer ref, check & fix
	prev = m[1][log2.DEBUG][1].WriterInd
	m[1][log2.DEBUG][1].WriterInd = 3
	_, err = logfu.New(
		[]logfu.FiltererFac{nil},
		[]logfu.SerializerFac{nil, nil},
		[]logfu.WriterFac{nil, nil},
		m,
		false)
	m[1][log2.DEBUG][1].WriterInd = prev
	if err == nil {
		t.Error("expected config failure 3")
	}

	// introduce unref'd filter & check for err
	_, err = logfu.New(
		[]logfu.FiltererFac{nil, nil},
		[]logfu.SerializerFac{nil, nil},
		[]logfu.WriterFac{nil, nil},
		m,
		false)
	if err == nil {
		t.Error("expected config failure 4")
	}

	// introduce unref'd serializer & check for err
	_, err = logfu.New(
		[]logfu.FiltererFac{nil},
		[]logfu.SerializerFac{nil, nil, nil},
		[]logfu.WriterFac{nil, nil},
		m,
		false)
	if err == nil {
		t.Error("expected config failure 5")
	}

	// introduce unref'd writer & check for err
	_, err = logfu.New(
		[]logfu.FiltererFac{nil},
		[]logfu.SerializerFac{nil, nil},
		[]logfu.WriterFac{nil, nil, nil},
		m,
		false)
	if err == nil {
		t.Error("expected config failure 6")
	}
}
