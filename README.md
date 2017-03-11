# logfu
--
    import "github.com/msample/logfu"

Implementations for github.com/msample/log2 LogFunc API supporting syslog,
fileio, JSON, logrotate, and stepping through different logging configurations
at runtime in response to OS signals.

Experimental work in progress. Incomplete. Sharp edges. Limited tests

Make sure Writers created by the writer factories are thread-safe per Write() to
ensure concurrent Write() calls to the same Writer will not have their output
interleaved.

Note: don't do log2 logging in the factories, filterers, serializers or writers
used with logfu.Config

## Usage

```go
const (
	SyslogPriority = syslog.LOG_INFO | syslog.LOG_LOCAL2 // all syslog usage gets this setup
)
```

#### func  CopyModes

```go
func CopyModes(m []Mode) []Mode
```
CopyModes returns deep copy of of given Mode slice. Useful if you want derive a
mode from another one.

#### func  FileWriterFac

```go
func FileWriterFac(filepath string) func() (io.Writer, error)
```
Returns a sync writer that append to the given file, creating the file if
necessary. FIXME expose file mode as param

#### func  IdentityFilter

```go
func IdentityFilter(keyvals []interface{}) ([]interface{}, error)
```

#### func  JSONSerialize

```go
func JSONSerialize(w io.Writer, kvs []interface{}) error
```
JSONSerialize is the JSON serializer implementation

#### func  LimitWriterFac

```go
func LimitWriterFac(f func() (io.Writer, error), maxSizePerWrite int) func() (io.Writer, error)
```

#### func  LogfmtSerialize

```go
func LogfmtSerialize(w io.Writer, kvs []interface{}) error
```
LogfmtSerialize is the log fmt serializer implementation

#### func  MultiWriterFac

```go
func MultiWriterFac(wfs ...func() (io.Writer, error)) func() (io.Writer, error)
```

#### func  RSWriterFac

```go
func RSWriterFac(f func() (io.Writer, error)) func() (io.Writer, error)
```
Wraps the given Writer to an add an ascii RS (record separator) before each
write and an LF after. Useful for producing json-seq when each individual write
to the wrapped writer is a JSON value.

#### func  StderrWriter

```go
func StderrWriter() (io.Writer, error)
```
Returns a sync Writer to os.Stderr

#### func  StdoutWriter

```go
func StdoutWriter() (io.Writer, error)
```
Returns a sync Writer to os.Stdout

#### func  SyncWriterFac

```go
func SyncWriterFac(f func() (io.Writer, error)) func() (io.Writer, error)
```
Returns a wrapper version of the given Writer that only permits one write call
at a time.

#### func  SyslogWriterFac

```go
func SyslogWriterFac() func() (io.Writer, error)
```
Returns a WriterFac that returns a thread-safe local syslog writer. Uses
SyslogPriority as the priority and "" as the tag (FIXME - expose syslog tag as
param?)

#### func  TCPSyslogWriterFac

```go
func TCPSyslogWriterFac(addr string) func() (io.Writer, error)
```
Returns a WriterFac returns a syslog TCP Writer to the given address. Addr
format is as per log/syslog.Dial().

#### func  UDPSyslogWriterFac

```go
func UDPSyslogWriterFac(addr string) func() (io.Writer, error)
```
Returns a WriterFac that uses syslog UDP protocol to the given address. Addr
format is as per log/syslog.Dial().

#### type Config

```go
type Config struct {
}
```


#### func  New

```go
func New(filtererFacs []FiltererFac,
	serializerFacs []SerializerFac,
	writerFacs []WriterFac,
	modes []Mode,
	recreateOnShift bool) (*Config, error)
```
New creates log configuration modes that can be stepped through to set the
log2.Info, Debug etc funcs. Change modes using the NextMode, PrevMode and
ChangeToMode methods. The Modes param defines the allowed configuration modes
only one of which is in effect at a given time. Each mode references a specific
combination of filterer, serializer and writer factories to be used to build the
log levels. Filterers are given the raw log parameters and may add or remove
keyavls, the resulting keyvals, if any, are serailized by the serializer to the
writer.

New returns a Config without applying any mode. Use ChangeToMode to set the
first logging mode.

Look at the example in logfu_test.go, this is a bit clunky

#### func (*Config) ChangeToMode

```go
func (o *Config) ChangeToMode(mode int, force, recreate bool) error
```
ChangeToMode changes to the given mode index. Does nothing if already in that
mode unless force is true.

Use this with force==true after first creating a Config to initialize the first
mode you want.

#### func (*Config) HomeMode

```go
func (o *Config) HomeMode() error
```
HomeMode loads modes[0] of the config. If already in home mode it does nothing

#### func (*Config) NextMode

```go
func (o *Config) NextMode() error
```
Shift to the next entry in the modes slice

#### func (*Config) PrevMode

```go
func (o *Config) PrevMode() error
```
Shift to the previous entry in the modes slice

#### func (*Config) ReloadMode

```go
func (o *Config) ReloadMode() error
```
Recreate the current log config by re-creating the filters, serializers and
writers and re-swapping them into their log2 levels (e.g. in response to HUP)

#### func (*Config) SignalControlOff

```go
func (o *Config) SignalControlOff()
```
SignalControlOff ceases changing log modes in response to signals

#### func (*Config) SignalControlOn

```go
func (o *Config) SignalControlOn()
```
SignalControlOn makes it so SIG_USR1 calls NextMode, SIG_USR2 calls HomeMode,
and SIG_HUP reloads the current log mode.

#### type FilterFunc

```go
type FilterFunc func([]interface{}) ([]interface{}, error)
```


#### func (FilterFunc) Filter

```go
func (o FilterFunc) Filter(keyvals []interface{}) ([]interface{}, error)
```

#### type Filterer

```go
type Filterer interface {
	Filter(inKeyvals []interface{}) (outKeyvals []interface{}, err error)
}
```

Filterer can add context to your log statements such as timestamp, or to mutate
or remove fields from the Log call and upstream filters. If the return value
from the Filter is zero length, the Log() call returns at that point (nothing is
logged).

#### func  IdentityFilterFac

```go
func IdentityFilterFac() (Filterer, error)
```

#### type FiltererFac

```go
type FiltererFac func() (Filterer, error)
```


#### type Fsw

```go
type Fsw struct {
	FilterInd     int
	SerializerInd int
	WriterInd     int
}
```

Fsw holds the FiltererFac, SerializerFac and WriterFac references that will be
used together to produce log output. Filter->Serialize->Write.

#### type LimitWriter

```go
type LimitWriter struct {
}
```


#### func (*LimitWriter) Write

```go
func (w *LimitWriter) Write(p []byte) (n int, err error)
```

#### type Mode

```go
type Mode map[log2.Level][]Fsw
```

Mode defines a logging configuration by specifying the log levels that are not
No-Ops in terms of a set of filter-serializer-writer tuples. Each log() call to
a level has its keyvals fed through each filterer-serializer-writer tuple for
that level (if any)

#### func  CopyMode

```go
func CopyMode(m Mode) Mode
```
CopyMode returns a deep copy of the given Mode

#### type MultiWriterCloser

```go
type MultiWriterCloser struct {
	io.Writer // wraps all writers for Write
}
```


#### func  NewMultiWriterCloser

```go
func NewMultiWriterCloser(w ...io.Writer) *MultiWriterCloser
```

#### func (*MultiWriterCloser) Close

```go
func (o *MultiWriterCloser) Close() error
```

#### type RSWriter

```go
type RSWriter struct {
}
```


#### func (*RSWriter) Write

```go
func (r *RSWriter) Write(p []byte) (n int, err error)
```

#### type Serializer

```go
type Serializer interface {
	Serialize(w io.Writer, inKeyvals []interface{}) error
}
```

Serializer converts a series of kv pairs to a single []byte and writes it to the
given Writer in a single Write call. Use a MultiWriter to avoid unecessary
re-serialization.

#### func  JSONSerializerFac

```go
func JSONSerializerFac() (Serializer, error)
```
JSONSerializerFac is a factory for JSON Serializers

#### func  LogfmtSerializerFac

```go
func LogfmtSerializerFac() (Serializer, error)
```
LogfmtSerializerFac is a factory for JSON Serializers

#### type SerializerFac

```go
type SerializerFac func() (Serializer, error)
```


#### type SerializerFunc

```go
type SerializerFunc func(io.Writer, []interface{}) error
```

SerializerFunc type allows you convert a serializer function to a Serializer
interface

#### func (SerializerFunc) Serialize

```go
func (o SerializerFunc) Serialize(w io.Writer, keyvals []interface{}) error
```

#### type WriterFac

```go
type WriterFac func() (io.Writer, error)
```
