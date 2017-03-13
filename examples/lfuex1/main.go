// lfuex1 uses a canned config and pflag commmand line setup with
// signal handling on.  The config has with 3 modes:
//
//  0 [error, warn, audit] to stdout & ./lfu_ex1.out
//  1 [audit] to stdout & ./lfu_ex1.out
//  2 [error, warn, audit, info, debug] to stdout & ./lfu_ex1.out
//
// lfuex1 runs until killed, running through a loop that tries each
// log level every 5 seconds. This should give you time to see the
// effect of sending it OS signals. It also shifts loggin modes on its
// own every 10 loop iterations.
//
// SIG_HUP will cause it to re-initialize the log setup so it will
// close the ./lfu_ex1.out file and re-open it for appending.  Try
// mv'ing the lfu_ex.out to lfu_ex.out.1, and either wait for the next
// mode switch or sig hup it.  It should stop writing to the moved
// file and create and start writing to a new lfu_ex.out
//
// SIG_USR1 will shift no the next log mode. SIG_USR2 will go to mode
// 0.
//
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/msample/log2"
	"github.com/msample/logfu/lib/lfucfg"
	"github.com/msample/logfu/lib/lfuflag"
)

func main() {
	help := flag.Bool("help", false, "print help and exit")
	lfuflag.AddDefaultLogfuFlags()
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	c := 0
	iters := 0
	cfg, err := lfucfg.FileStdOutJsonWithSigs("lfu_ex1.out")
	if err != nil {
		fmt.Printf("Exiting, lfucfg.FileStdOutJsonWithSigs is having problems: %v\n", err)
		os.Exit(1)
	}
	for {
		fmt.Printf("iter: %v\n", iters)
		err := log2.Debug("msg", "starting loop (Debug)", "count", c)
		if err != nil {
			fmt.Printf("log err: %v\n", err)
		}
		c++

		log2.Info("msg", "2nd log call in loop (Info)", "count", c)
		c++

		log2.Warn("msg", "3rd  log call in loop (Warn)", "count", c)
		c++

		log2.Error("msg", "4th  log call in loop (Error)", "count", c)
		c++

		log2.Audit("msg", "5th &last log call in loop (Audit)", "count", c)
		c++

		time.Sleep(5 * time.Second)

		iters++
		if iters%10 == 0 {
			log2.Audit("msg", "Shifting to next logging mode")
			cfg.NextMode()
		}
	}
}
