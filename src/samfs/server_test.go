package samfs

import (
	"flag"
	"os"
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

var TestCtx *testContext

func TestMain(m *testing.M) {
	flag.Parse()

	// setup
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
	os.Exit(exitCode)
}

func setup() (*testContext, error) {
	tCtx := &testContext{}

	// setup samfs server
	s, serr := NewServer(".")
	if serr != nil {
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
	t.Run("Mount", func(t *testing.T) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		// TODO(mihir): check return value of mount as well
		_, err := TestCtx.Client.Mount(ctx, &pb.MountRequest{RootDirectory: "/"})
		if err != nil {
			t.Fatalf("mounting failed with error :: %s", err.Error())
			t.Fail()
		}
	})
}
