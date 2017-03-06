// Implementations for github.com/msample/log2 LogFunc API supporting
// syslog, fileio, JSON, logrotate, and stepping through different
// logging configurations at runtime in response to OS signals.
//
// Experimental work in progress. Incomplete. Sharp edges. Limited tests
//
// Make sure Writers created by the writer factories are thread-safe
// per Write() to ensure concurrent Write() calls to the same Writer
// will not have their output interleaved.
//
// Note: don't do log2 logging in the factories, filterers, serializers or
// writers used with logfu.Config
//
//
//
package logfu

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	syslog "github.com/RackSec/srslog" // more standards compliant than log/sylog
	"github.com/msample/log2"
)

const (
	SyslogPriority = syslog.LOG_INFO | syslog.LOG_LOCAL2 // all syslog usage gets this setup
)

type WriterFac func() (io.Writer, error)
type FiltererFac func() (Filterer, error)
type SerializerFac func() (Serializer, error)

// Filterer can add context to your log statements such as timestamp,
// or to mutate or remove fields from the Log call and upstream
// filters.  If the return value from the Filter is zero length, the
// Log() call returns at that point (nothing is logged).
type Filterer interface {
	Filter(inKeyvals []interface{}) (outKeyvals []interface{}, err error)
}

// Serializer converts a series of kv pairs to a single []byte and
// writes it to the given Writer in a single Write call.  Use a
// MultiWriter to avoid unecessary re-serialization.
type Serializer interface {
	Serialize(w io.Writer, inKeyvals []interface{}) error
}

type Config struct {
	mutex           sync.Mutex
	filtererFacs    []FiltererFac
	serializerFacs  []SerializerFac
	writerFacs      []WriterFac
	modes           []Mode
	recreateOnShift bool
	currMode        int
	modeVals        *modeVals
	sigCh           chan os.Signal
	sigStopCh       chan struct{}
}

type Mode map[int][]Fsw

// Fsw holds the FiltererFac, SerializerFac and WriterFac references
// that will be used together to produce log
// output. Filter->Serialize->Write.
type Fsw struct {
	FilterInd     int
	SerializerInd int
	WriterInd     int
}

func New(filtererFacs []FiltererFac,
	serializerFacs []SerializerFac,
	writerFacs []WriterFac,
	modes []Mode,
	recreateOnShift bool) (*Config, error) {

	// Verify that:
	//   all factories are referenced
	//   all refrences are inbounds
	//   fsw slices are non-empty
	refdFiltererFacs := make(map[int]bool)
	refdSerailizerFacs := make(map[int]bool)
	refdWriterFacs := make(map[int]bool)
	for i := range modes {
		for _, v := range modes[i] {
			if len(v) == 0 {
				return nil, fmt.Errorf("Fsw slice may not be empty")
			}
			for _, f := range v {
				refdFiltererFacs[f.FilterInd] = true
				refdSerailizerFacs[f.SerializerInd] = true
				refdWriterFacs[f.WriterInd] = true
				if f.FilterInd >= len(filtererFacs) || f.FilterInd < 0 {
					return nil, fmt.Errorf("FilterFac reference %v is out of range(%v)", f.FilterInd, len(filtererFacs))
				}
				if f.SerializerInd >= len(serializerFacs) || f.SerializerInd < 0 {
					return nil, fmt.Errorf("SerializerFac reference %v is out of range(%v)", f.SerializerInd, len(serializerFacs))
				}
				if f.WriterInd >= len(writerFacs) || f.WriterInd < 0 {
					return nil, fmt.Errorf("WriterFac reference %v is out of range(%v)", f.WriterInd, len(writerFacs))
				}
			}
		}
	}
	for i := 0; i < len(filtererFacs); i++ {
		if !refdFiltererFacs[i] {
			return nil, fmt.Errorf("unreferenced filterFacs entry: %v", i)
		}
	}
	for i := 0; i < len(serializerFacs); i++ {
		if !refdSerailizerFacs[i] {
			return nil, fmt.Errorf("unreferenced serializerFacs entry: %v", i)
		}
	}
	for i := 0; i < len(writerFacs); i++ {
		if !refdWriterFacs[i] {
			return nil, fmt.Errorf("unreferenced writerFacs entry: %v", i)
		}
	}

	// copy input vals to prohibit external mutation
	ff := make([]FiltererFac, len(filtererFacs))
	copy(ff, filtererFacs)
	sf := make([]SerializerFac, len(serializerFacs))
	copy(sf, serializerFacs)
	wf := make([]WriterFac, len(writerFacs))
	copy(wf, writerFacs)
	rv := &Config{
		filtererFacs:    ff,
		serializerFacs:  sf,
		writerFacs:      wf,
		modes:           CopyModes(modes),
		recreateOnShift: recreateOnShift,
	}
	rv.modeVals = rv.newModeVals()
	return rv, nil
}

// make it so SIG_USR1 calls NextMode, SIG_USR2 calls HomeMode, and
// SIG_HUP reloads the current log mode.
func (o *Config) SignalControlOn() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.sigCh = make(chan os.Signal, 32)
	o.sigStopCh = make(chan struct{})
	signal.Notify(o.sigCh, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP)
	go func(sigCh <-chan os.Signal, doneCh <-chan struct{}) {
		for {
			select {
			case <-doneCh:
				return
			case s := <-sigCh:
				switch s {
				case (syscall.SIGUSR1):
					o.NextMode()

				case (syscall.SIGUSR2):
					o.HomeMode()

				case (syscall.SIGHUP):
					o.ReloadMode()
				}
			}
		}
	}(o.sigCh, o.sigStopCh)
}

func (o *Config) SignalControlOff() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	signal.Stop(o.sigCh)
	close(o.sigStopCh)
	o.sigCh = nil
	o.sigStopCh = nil
}

// Recreate the current log config by re-creating the filters,
// serializers and writers and re-swapping them into their log2 levels
func (o *Config) ReloadMode() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.changeToMode(o.currMode, true, true)
}

// Shift to the next entry in the modes slice
func (o *Config) NextMode() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	next := o.currMode + 1
	if next >= len(o.modes) {
		next = 0
	}
	return o.changeToMode(next, false, o.recreateOnShift)
}

// Shift to the previous entry in the modes slice
func (o *Config) PrevMode() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	prev := o.currMode - 1
	if prev < 0 {
		prev = len(o.modes) - 1
	}
	return o.changeToMode(prev, false, o.recreateOnShift)
}

// HomeMode loads modes[0] of the config. If already in home mode it
// does nothing
func (o *Config) HomeMode() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.changeToMode(0, false, o.recreateOnShift)
}

// ChangeToMode changes to the given mode index. Does nothing if
// already in that mode unless force is true.
func (o *Config) ChangeToMode(mode int, force, recreate bool) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return o.changeToMode(mode, force, recreate)
}

// must hold o.mutex to call this
func (o *Config) changeToMode(mode int, force, recreate bool) error {

	if mode > len(o.modes) || mode < 0 {
		return fmt.Errorf("ChangeToMode: mode out of range: %v", mode)
	}
	if o.currMode == mode && !force {
		return nil
	}
	mv, toClose, err := o.modeValsForMode(mode, recreate)

	if err != nil {
		return err
	}

	// swap log funcs to use new modeVals
	o.currMode = mode
	o.modeVals = mv
	for k, v := range o.modes[mode] {
		// consider providing mutex in log2 to used when
		// swapping more then one func
		log2.Swap(k, makeLogFunc(mv, v))
	}

	// swap nop log in for ones not replaced by this mode
	swapNop(o.modes[mode])

	// now close anything not reused from the previous mode.
	// except stderr and stdout since you probably don't want to
	// close those.
	for _, c := range toClose {
		if f, ok := c.(*os.File); ok {
			if f == os.Stderr || f == os.Stdout {
				continue
			}
		}
		err := c.Close()
		if err != nil {
			fmt.Printf("logfu: Could not close writer, serializer or filter when changing mode. Resource type: %T", c)
		}
	}

	return nil
}

func CopyModes(m []Mode) []Mode {
	rv := make([]Mode, 0, len(m))
	for i := range m {
		rv = append(rv, CopyMode(m[i]))
	}
	return rv
}

func CopyMode(m Mode) Mode {
	rv := make(Mode)
	for k, v := range m {
		f := make([]Fsw, len(v))
		copy(f, v)
		rv[k] = f
	}
	return rv
}

func (o *Config) newModeVals() *modeVals {
	return newModeVals(len(o.filtererFacs),
		len(o.serializerFacs), len(o.writerFacs))
}

func (o *Config) modeValsForMode(mode int, recreate bool) (*modeVals, []io.Closer, error) {
	rv := o.newModeVals()
	ffm := o.filtersForMode(mode)
	sfm := o.serializersForMode(mode)
	wfm := o.writersForMode(mode)
	var toClose []io.Closer
	var err error
	for k, v := range ffm {
		if v {
			f := o.modeVals.filters[k]
			if recreate || f == nil {
				if c, ok := f.(io.Closer); ok {
					toClose = append(toClose, c)
				}
				f, err = o.filtererFacs[k]()
				if err != nil {
					return nil, nil, err
				}
			}
			rv.filters[k] = f
		}
	}
	for k, v := range sfm {
		if v {
			f := o.modeVals.serializers[k]
			if recreate || f == nil {
				if c, ok := f.(io.Closer); ok {
					toClose = append(toClose, c)
				}
				f, err = o.serializerFacs[k]()
				if err != nil {
					return nil, nil, err
				}
			}
			rv.serializers[k] = f
		}
	}
	for k, v := range wfm {
		if v {
			f := o.modeVals.writers[k]
			if recreate || f == nil {
				if c, ok := f.(io.Closer); ok {
					toClose = append(toClose, c)
				}
				f, err = o.writerFacs[k]()
				if err != nil {
					return nil, nil, err
				}
			}
			rv.writers[k] = f
		}
	}

	// find things that need to be closed from the prev mode that
	// weren't already added to toClose or re-used in the new mode
	// above
	for i, f := range o.modeVals.filters {
		if !ffm[i] {
			if c, ok := f.(io.Closer); ok {
				toClose = append(toClose, c)
			}
		}
	}
	for i, s := range o.modeVals.serializers {
		if !sfm[i] {
			if c, ok := s.(io.Closer); ok {
				toClose = append(toClose, c)
			}
		}
	}
	for i, w := range o.modeVals.writers {
		if !wfm[i] {
			if c, ok := w.(io.Closer); ok {
				toClose = append(toClose, c)
			}
		}
	}
	return rv, toClose, nil
}

// returns list of filter indexes used in the given mode
func (o *Config) filtersForMode(mode int) map[int]bool {
	rv := make(map[int]bool)
	for _, v := range o.modes[mode] {
		for i := range v {
			rv[v[i].FilterInd] = true
		}
	}
	return rv
}

// returns list of serializer indexes used in the given mode
func (o *Config) serializersForMode(mode int) map[int]bool {
	rv := make(map[int]bool)
	for _, v := range o.modes[mode] {
		for i := range v {
			rv[v[i].SerializerInd] = true
		}
	}
	return rv
}

// returns list of writer indexes used in the given mode
func (o *Config) writersForMode(mode int) map[int]bool {
	rv := make(map[int]bool)
	for _, v := range o.modes[mode] {
		for i := range v {
			rv[v[i].WriterInd] = true
		}
	}
	return rv
}

// put Nop Log func in all level that are NOT in the given mode
func swapNop(s Mode) {
	for i := 0; i < int(log2.HIGHEST); i++ {
		if s[i] == nil {
			log2.Swap(i, nil)
		}
	}
}

func makeLogFunc(mv *modeVals, c []Fsw) log2.LogFunc {
	mv2 := mv.copy() // func created below binds the copies
	c2 := make([]Fsw, len(c))
	copy(c2, c)

	filts := make(map[int]bool)
	for i := range c2 {
		filts[c2[i].FilterInd] = true
	}
	oneFilt := len(filts) > 0

	if oneFilt {
		// optimized version with a single shared filter. Win
		// here is that if you have a timestamp generator
		// added by a filter, all the serializers & writers get
		// the same value.  MultiWriter can be used by callers
		// gives simlar re-use of a serialzed result
		return func(keyvals ...interface{}) error {
			kv, err := mv2.filters[c2[0].FilterInd].Filter(keyvals)
			if err != nil {
				return err
			}
			if len(kv) == 0 {
				return nil // filtered out everything
			}
			for i := range c2 {
				w := mv2.writers[c2[i].WriterInd]
				err = mv2.serializers[c2[i].SerializerInd].Serialize(w, kv)
				if err != nil {
					return err
				}
			}
			return nil
		}

	}

	return func(keyvals ...interface{}) error {
		for i := range c2 {
			kv, err := mv2.filters[c2[i].FilterInd].Filter(keyvals)
			if err != nil {
				return err
			}
			if len(kv) == 0 {
				return nil // filtered out everything
			}
			w := mv2.writers[c2[i].WriterInd]
			err = mv2.serializers[c2[i].SerializerInd].Serialize(w, kv)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

type FilterFunc func([]interface{}) ([]interface{}, error)

func (o FilterFunc) Filter(keyvals []interface{}) ([]interface{}, error) {
	return o(keyvals)
}

type SerializerFunc func(io.Writer, []interface{}) error

func (o SerializerFunc) Serialize(w io.Writer, keyvals []interface{}) error {
	return o(w, keyvals)
}

// modeVals holds the objects created from the factories for the current mode
type modeVals struct {
	filters     []Filterer   // sparse, always len(Config.filterFacs), may have nil entries
	serializers []Serializer // sparse, always len(Config.serializerFacs), may have nil entries
	writers     []io.Writer  // sparse, always len(Config.writerFacs), may have nil entries
}

func newModeVals(flen, slen, wlen int) *modeVals {
	return &modeVals{
		make([]Filterer, flen),
		make([]Serializer, slen),
		make([]io.Writer, wlen),
	}
}

func (o *modeVals) copy() *modeVals {
	f := make([]Filterer, len(o.filters))
	copy(f, o.filters)
	s := make([]Serializer, len(o.serializers))
	copy(s, o.serializers)
	w := make([]io.Writer, len(o.writers))
	copy(w, o.writers)
	return &modeVals{filters: f, serializers: s, writers: w}
}
