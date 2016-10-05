package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: samfs-server -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	flag.Parse()
}

func main() {
	e := errors.New("samfs server stub")
	glog.Errorf(e.Error())
}
