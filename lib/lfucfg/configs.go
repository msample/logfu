// lfucfg provides convenient functions to build and initialize
// common logging configurations.
//
package lfucfg

import (
	"github.com/msample/log2"
	"github.com/msample/logfu"
)

// FileStdOutJsonWithSigs sets up and returns a logfu Config object
// with OS signal control enabled and 3 modes:
//
// 0 - Error,Warn & Audit go to stdout and the named file
// 1 - Audit goes to stdout and the named file
// 2 - Error,Warn,Warn,Debug, & Audit go to stdout and the named file
//
// Forces changes to mode zero before returning so no need to apply it
// yourself.
func FileStdOutJsonWithSigs(filename string) (*logfu.Config, error) {
	m := []logfu.Mode{
		{
			log2.ERROR: []logfu.Fsw{{0, 0, 0}},
			log2.WARN:  []logfu.Fsw{{0, 0, 0}},
			log2.AUDIT: []logfu.Fsw{{0, 0, 0}},
		},
		{
			log2.AUDIT: []logfu.Fsw{{0, 0, 0}},
		},
		{
			log2.ERROR: []logfu.Fsw{{0, 0, 0}},
			log2.WARN:  []logfu.Fsw{{0, 0, 0}},
			log2.DEBUG: []logfu.Fsw{{0, 0, 0}},
			log2.INFO:  []logfu.Fsw{{0, 0, 0}},
			log2.AUDIT: []logfu.Fsw{{0, 0, 0}},
		},
	}

	// since identity filter and json serializer are common to both
	// writers, the stdout and the file can be wrapped in a
	// MultiWriterFac
	rv, err := logfu.New(
		[]logfu.FiltererFac{logfu.IdentityFilterFac},
		[]logfu.SerializerFac{logfu.JSONSerializerFac},
		[]logfu.WriterFac{logfu.MultiWriterFac(logfu.StdoutWriter,
			logfu.FileWriterFac(filename))},
		m,
		true)
	if err != nil {
		return nil, err
	}
	rv.ChangeToMode(0, true, true)
	return rv, nil
}
