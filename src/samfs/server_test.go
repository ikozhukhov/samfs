package samfs

import (
	"flag"
	"os"
	"os/exec"
	"path"
	"sync"
	"testing"
	"time"

	pb "github.com/smihir/samfs/src/proto"

	//"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type testContext struct {
	Client     pb.NFSClient
	Server     *SamFSServer
	ClientConn *grpc.ClientConn
	wg         sync.WaitGroup
}

const mountDir = "samfs_testdir"

var TestCtx *testContext

func TestMain(m *testing.M) {
	flag.Parse()

	// setup
	wd, werr := os.Getwd()
	if werr != nil {
		panic(werr.Error())
	}
	tCtx, err := setup()
	if err != nil {
		panic(err.Error())
	}
	TestCtx = tCtx

	exitCode := m.Run()

	// teardown
	tCtx.ClientConn.Close()
	tCtx.Server.Stop()
	tCtx.wg.Wait()
	if err := os.Remove(path.Join(wd, mountDir)); err != nil {
		panic(err.Error())
	}
	os.Exit(exitCode)
}

func setup() (*testContext, error) {
	tCtx := &testContext{}
	ok := true

	wd, werr := os.Getwd()
	if werr != nil {
		return nil, werr
	}

	// setup samfs server
	merr := os.Mkdir(path.Join(wd, mountDir), 0777)
	if merr != nil {
		return nil, merr
	}
	defer func() {
		if ok == false {
			os.Remove(path.Join(wd, mountDir))
		}
	}()

	s, serr := NewServer(path.Join(wd, mountDir))
	if serr != nil {
		ok = false
		return nil, serr
	}
	tCtx.Server = s

	// run server
	tCtx.wg.Add(1)
	go func() {
		tCtx.Server.Run()
		tCtx.wg.Done()
	}()

	// setup grpc client to talk to samfs server
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := grpc.DialContext(ctx, "127.0.0.1:24100",
		grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	tCtx.Client = pb.NewNFSClient(conn)
	tCtx.ClientConn = conn

	return tCtx, nil
}

func TestSamfs(t *testing.T) {
	var rootFh, innerFh *pb.FileHandle
	wd, werr := os.Getwd()
	if werr != nil {
		t.Fatalf("failed to get working directory :: %s", werr.Error())
		t.Fail()
	}

	md := path.Join(wd, mountDir)
	t.Run("Mount", func(t *testing.T) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		resp, err := TestCtx.Client.Mount(ctx, &pb.MountRequest{RootDirectory: "/"})
		if err != nil {
			t.Fatalf("mounting failed with error :: %s", err.Error())
			t.Fail()
		}
		rootFh = resp.FileHandle
	})

	t.Run("Mkdir", func(t *testing.T) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		req := &pb.LocalDirectoryRequest{
			DirectoryFileHandle: rootFh,
			Name:                "innerdir",
		}
		resp, err := TestCtx.Client.Mkdir(ctx, req)
		if err != nil {
			cmd := exec.Command("tree")
			out, _ := cmd.CombinedOutput()
			t.Error(string(out))
			t.Fatalf("mkdir failed with error :: %s", err.Error())
			t.Fail()
		}
		innerFh = resp.FileHandle

		// check if directory is actually created
		directoryPath := path.Join(md, "innerdir")
		if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
			t.Error(err.Error())
			cmd := exec.Command("tree")
			out, _ := cmd.CombinedOutput()
			t.Error(string(out))
			t.Fatalf("mkdir did not create a directory %s", directoryPath)
			t.Fail()
		}
	})

	t.Run("InodeGenerationNumbers", func(t *testing.T) {
		i, g, err := TestCtx.Server.GetInodeAndGenerationNumbers(md)
		if err != nil {
			t.Fatalf("failed to get inode and generation numbers for %s :: %s",
				md, err.Error())
			t.Fail()
		} else {
			t.Logf("inode: %d, generation number: %d", i, g)
		}
	})

	t.Run("Rmdir", func(t *testing.T) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		// according to the spec. fh should be of parent of the sub-directory
		// and Name should be the name of the sub-directory to be removed.
		req := &pb.LocalDirectoryRequest{
			DirectoryFileHandle: rootFh,
			Name:                "innerdir",
		}
		_, err := TestCtx.Client.Rmdir(ctx, req)
		if err != nil {
			t.Fatalf("rmdir failed with error :: %s", err.Error())
			t.Fail()
		}

		// check if directory is actually removed
		directoryPath := path.Join(md, "innerDir")
		if _, err := os.Stat(directoryPath); os.IsExist(err) {
			cmd := exec.Command("tree")
			out, _ := cmd.CombinedOutput()
			t.Error(string(out))
			t.Fatalf("rmdir did not remove a directory %s", directoryPath)
			t.Fail()
		}
	})
}
