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

#### func (*Config) ChangeToMode

```go
func (o *Config) ChangeToMode(mode int, force, recreate bool) error
```
ChangeToMode changes to the given mode index. Does nothing if already in that
mode unless force is true.

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
writers and re-swapping them into their log2 levels

#### func (*Config) SignalControlOff

```go
func (o *Config) SignalControlOff()
```

#### func (*Config) SignalControlOn

```go
func (o *Config) SignalControlOn()
```
make it so SIG_USR1 calls NextMode, SIG_USR2 calls HomeMode, and SIG_HUP reloads
the current log mode.

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

#### type Mode

```go
type Mode map[int][]Fsw
```


#### func  CopyMode

```go
func CopyMode(m Mode) Mode
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

#### type SerializerFac

```go
type SerializerFac func() (Serializer, error)
```


#### type SerializerFunc

```go
type SerializerFunc func(io.Writer, []interface{}) error
```


#### func (SerializerFunc) Serialize

```go
func (o SerializerFunc) Serialize(w io.Writer, keyvals []interface{}) error
```

#### type WriterFac

```go
type WriterFac func() (io.Writer, error)
```
