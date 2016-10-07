package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/smihir/samfs/src/samfs"
)

var rootDirectory *string

func usage() {
	fmt.Fprintf(os.Stderr, "usage: samfs-server -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	rootDirectory = flag.String("root", "default root", "this is root of the FS")
	flag.Parse()
}

func main() {
	if rootDirectory == nil {
		usage()
	}
	_, _ = samfs.NewServer(*rootDirectory)
	e := errors.New("samfs server stub")
	glog.Errorf(e.Error())
}
