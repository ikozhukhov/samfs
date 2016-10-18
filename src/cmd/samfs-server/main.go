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
	rootDirectory *string
	port          *string
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: samfs-server -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	rootDirectory = flag.String("root", "", "this is root of the FS")
	port = flag.String("port", "24100", "this is port of communication")
	flag.Parse()
}

func main() {
	if *rootDirectory == "" {
		usage()
	}
	s, _ := samfs.NewServer(*rootDirectory, *port)
	s.Run()
	e := errors.New("samfs server stub")
	glog.Errorf(e.Error())
}
