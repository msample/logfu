package logfu

import (
	"fmt"
	"io"
	"os"

	syslog "github.com/RackSec/srslog" // more standards compliant than log/sylog
	"github.com/go-kit/kit/log"
)

const (
	SyslogPriority = syslog.LOG_INFO | syslog.LOG_LOCAL2 // all syslog usage gets this setup
)

// Returns a sync Writer to os.Stderr
func StderrWriter() (io.Writer, error) {
	return log.NewSyncWriter(os.Stderr), nil
}

// Returns a sync Writer to os.Stdout
func StdoutWriter() (io.Writer, error) {
	return log.NewSyncWriter(os.Stdout), nil
}

// Returns a sync writer that append to the given file, creating the
// file if necessary. FIXME expose file mode as param
func FileWriterFac(filepath string) func() (io.Writer, error) {
	return func() (io.Writer, error) {
		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
}

// Returns a WriterFac that returns a thread-safe local syslog
// writer. Uses SyslogPriority as the priority and "" as the tag
// (FIXME - expose syslog tag as param?)
func SyslogWriterFac() func() (io.Writer, error) {
	return func() (io.Writer, error) {
		w, err := syslog.New(SyslogPriority, "")
		if err != nil {
			return nil, err
		}
		setFmt(w)
		return w, nil
	}
}

// Returns a WriterFac that uses syslog UDP protocol to the given
// address.  Addr format is as per log/syslog.Dial().
func UDPSyslogWriterFac(addr string) func() (io.Writer, error) {
	return func() (io.Writer, error) {
		w, err := syslog.Dial("udp", addr, SyslogPriority, "")
		if err != nil {
			return nil, err
		}
		setFmt(w)
		return w, nil
	}
}

// Returns a WriterFac returns a syslog TCP Writer to the given
// address.  Addr format is as per log/syslog.Dial().
func TCPSyslogWriterFac(addr string) func() (io.Writer, error) {
	return func() (io.Writer, error) {
		// syslog writer has an internal mutex to make writes thread safe
		// so no need for a SyncWriter wrapper
		w, err := syslog.Dial("tcp", addr, SyslogPriority, "")
		if err != nil {
			return nil, err
		}
		setFmt(w)
		return w, nil
	}
}

// single place were we choose Syslog protocol version
func setFmt(w *syslog.Writer) {
	w.SetFormatter(syslog.RFC3164Formatter)
}

// Returns a wrapper version of the given Writer that only permits one
// write call at a time.
func SyncWriterFac(f func() (io.Writer, error)) func() (io.Writer, error) {
	return func() (io.Writer, error) {
		w, err := f()
		if err != nil {
			return nil, err
		}
		return log.NewSyncWriter(w), nil
	}
}

func MultiWriterFac(wfs ...func() (io.Writer, error)) func() (io.Writer, error) {
	return func() (io.Writer, error) {
		var ws []io.Writer
		for _, wf := range wfs {
			w, err := wf()
			if err != nil {
				return nil, err
			}
			ws = append(ws, w)
		}
		return NewMultiWriterCloser(ws...), nil
	}
}

type MultiWriterCloser struct {
	io.Writer             // wraps all writers for Write
	writers   []io.Writer // maybe closers too
}

func NewMultiWriterCloser(w ...io.Writer) *MultiWriterCloser {
	//return &MultiWriterCloser{multiW: io.MultiWriter(w...), writers: w}
	return &MultiWriterCloser{io.MultiWriter(w...), w}
}

func (o *MultiWriterCloser) Close() error {
	var errs []error
	for _, w := range o.writers {
		if c, ok := w.(io.Closer); ok {
			err := c.Close()
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf("Error(s) closing MultiWriterCloser: %v\n", errs)
	}
	return nil
}

type LimitWriter struct {
	maxSize int
	w       io.Writer
}

func (w *LimitWriter) Write(p []byte) (n int, err error) {
	if len(p) > w.maxSize {
		return w.w.Write(p[:w.maxSize])
	}
	return w.w.Write(p)
}

func LimitWriterFac(f func() (io.Writer, error), maxSizePerWrite int) func() (io.Writer, error) {
	return func() (io.Writer, error) {
		w, err := f()
		if err != nil {
			return nil, err
		}
		return &LimitWriter{maxSize: maxSizePerWrite, w: w}, nil
	}
}

type RSWriter struct {
	w io.Writer
}

func (r *RSWriter) Write(p []byte) (n int, err error) {
	var i, j int
	n, err = r.w.Write([]byte{10}) // LF
	if err != nil {
		return
	}
	i, err = r.w.Write(p)
	n += i
	if err != nil {
		return
	}
	j, err = r.w.Write([]byte{30}) // RS
	n += j
	return
}

// Wraps the given Writer to an add an ascii RS (record separator)
// before each write and an LF after.  Useful for producing json-seq
// when each individual write to the wrapped writer is a JSON value.
func RSWriterFac(f func() (io.Writer, error)) func() (io.Writer, error) {
	return func() (io.Writer, error) {
		w, err := f()
		if err != nil {
			return nil, err
		}
		return &RSWriter{w}, nil
	}
}
