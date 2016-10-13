package main

import (
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
	mountDir := flag.String("mount", "test", "mount directory")
	flag.Parse()
	client, err := samfs.NewClient(server, port, mountDir)
	if err != nil {
		glog.Errorf("connection failed : %s", err.Error())
	}
	client.Run()
}
