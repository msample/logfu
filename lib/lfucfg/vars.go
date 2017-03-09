package lfucfg

// define vars used by config recipies so they can be set by flag,
// cobra etc

var (
	SyslogHost string
	SyslogPort int
	SyslogType string
)

// fixme add funcs to validate vars
