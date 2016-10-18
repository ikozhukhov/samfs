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
		server := "127.0.0.1"
		port := "24100"
		dir := "/trash"
		_, _ = samfs.NewClient(&server, &port, &dir)
	} else {
		_, _ = samfs.NewServer(".", "24100")
	}

	e := errors.New("samfs server stub")
	glog.Errorf(e.Error())
}
