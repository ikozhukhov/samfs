package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"golang.org/x/sys/unix"

	"github.com/golang/glog"
)

const niterations int = 10

var directory *string

func usage() {
	fmt.Fprintf(os.Stderr, "usage: read_test -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	directory = flag.String("dir", "", "this is the directory files are write to")
	flag.Parse()
}

func main() {
	if *directory == "" {
		usage()
	}

	var time1 unix.Timespec
	var time2 unix.Timespec

	for granularity := 16; granularity <= 16*1024*1024; granularity *= 2 {
		filename := fmt.Sprintf("%d.txt", granularity)
		filePath := path.Join(*directory, filename)
		fd, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0766)
		if err != nil {
			glog.Errorf("failed to open %s :: %v", filePath, err)
			return
		}

		data := make([]byte, granularity, granularity)
		times := make([]float64, niterations, niterations)

		for i := 0; i < niterations; i++ {
			unix.ClockGettime(unix.CLOCK_REALTIME, &time1)
			_, err = fd.Read(data)
			if err != nil {
				glog.Errorf("failed to read in iteration %d for granularity %d\n", i,
					granularity)
			}
			unix.ClockGettime(unix.CLOCK_REALTIME, &time2)
			nsecDiff := float64(unix.TimespecToNsec(time2) - unix.TimespecToNsec(time1))
			times[i] = nsecDiff
		}

		err = fd.Close()
		if err != nil {
			glog.Errorf("NOT EXITING! but failed to close %s :: %v", filename, err)
		}

		sum := float64(0)
		for i := 1; i < niterations; i++ {
			sum += times[i]
		}

		iter := float64(niterations - 1)
		fmt.Printf("granularity %d, %1.2f ns\n", granularity, sum/iter)
	}
}
