package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/smihir/samfs/src/samfs"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: samfsc -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	server := flag.String("server", "127.0.0.1", "server IP or name")
	port := flag.String("port", "24100", "server port")
	flag.Parse()
	_, err := samfs.NewClient(server, port)
	if err != nil {
		glog.Errorf("connection failed : %s", err.Error())
	}
	e := errors.New("samfs client stub")
	glog.Errorf(e.Error())
}
