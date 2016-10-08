package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"

	"github.com/smihir/samfs/src/samfs"
)

var (
	isclient = flag.Bool("c", false, "run client")
	isserver = flag.Bool("s", false, "run server")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: samfs <-cs>\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	flag.Parse()
}

func main() {
	if *isclient == false && *isserver == false {
		flag.Usage()
	} else if *isclient == true {
		_, _ = samfs.NewClient()
	} else {
		_, _ = samfs.NewServer(".")
	}

	e := errors.New("samfs server stub")
	glog.Errorf(e.Error())
}
