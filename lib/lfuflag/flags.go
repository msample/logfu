// lfuflag offers functions to up up some logfu command line
// flags. Call before flags.Parse(). Cobra elsewhere.  Does not use
// import/init() sideffect style in order to allow you to set flag
// default values.
//
// The flags defined here:
//
//  lfu_syslog_port
//	lfu_syslog_host
//	lfu_syslog_type

package lfuflag

import (
	"flag"

	"github.com/msample/logfu/lib/lfucfg"
)

type Defaults struct {
	SyslogPort      int
	SyslogHost      string
	SyslogTransport string
}

var ReasonableDefaults = Defaults{
	5514,
	"localhost",
	"udp",
}

// AddLogfuFlags adds command line flags that will initialze the
// exported vars forom logfu/lib/lfucfg.  These exported vars may be
// used by the lfucfg Config creation convenience function.  THe
// Default value will be used if the user does not explicilty provide
// that flag on the command line.
func AddLogfuFlags(d Defaults) {
	flag.IntVar(&lfucfg.SyslogPort, "lfu_syslog_port", d.SyslogPort, "syslog port")
	flag.StringVar(&lfucfg.SyslogHost, "lfu_syslog_host", d.SyslogHost, "syslog hostname or IP")
	flag.StringVar(&lfucfg.SyslogType, "lfu_syslog_type", d.SyslogTransport, "'udp' or 'tcp'")
}

// AddDefaultLogfuFlags calls AddLogfuFlags with the ReasonableDefault
// default values
func AddDefaultLogfuFlags() {
	AddLogfuFlags(ReasonableDefaults)
}
