package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"

	"samfs/src/samfs"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: samfsc -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	flag.Parse()
}

func main() {
	_, _ = samfs.NewClient()
	e := errors.New("samfs client stub")
	glog.Errorf(e.Error())
}
