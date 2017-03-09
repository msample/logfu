// fuflag sets up up some logfu command line flags. Just import this
// package somewhere to enable it if you use the "flag" pkg. Corbra
// elsewhere.
//
// The flags def
package lfuflag

import (
	"flag"

	"github.com/msample/logfu/lib/lfucfg"
)

func init() {
	flag.IntVar(&lfucfg.SyslogPort, "lfu_syslog_port", 5514, "syslog port")
	flag.StringVar(&lfucfg.SyslogHost, "lfu_syslog_host", "localhost", "syslog hostname or IP")
	flag.StringVar(&lfucfg.SyslogType, "lfu_syslog_type", "udp", "'udp' or 'tcp'")
}
