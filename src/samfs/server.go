package samfs

import (
	"os"

	"github.com/golang/glog"
	pb "github.com/smihir/samfs/src/proto"
	"golang.org/x/net/context"
)

type SamFSServer struct {
	rootDirectory string
}

var _ pb.NFSServer = &SamFSServer{}

func NewServer(rootDirectory string) (*SamFSServer, error) {
	if _, err := os.Stat(rootDirectory); os.IsNotExist(err) {
		glog.Errorf("path %s does not exist :: %v\n", rootDirectory, err)
		return nil, err
	}

	glog.Infof("FS root = %s\n", rootDirectory)

	s := &SamFSServer{
		rootDirectory: rootDirectory,
	}

	return s, nil
}

func (s *SamFSServer) Mount(ctx context.Context, req *pb.MountRequest) (*pb.FileHandleReply, error) {
	return nil, nil
}

func (s *SamFSServer) Lookup(ctx context.Context, req *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
	return nil, nil
}

func (s *SamFSServer) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadReply, error) {
	return nil, nil
}

func (s *SamFSServer) Write(ctx context.Context, req *pb.WriteRequest) (*pb.StatusReply, error) {
	return nil, nil
}

func (s *SamFSServer) Commit(ctx context.Context, req *pb.CommitRequest) (*pb.StatusReply, error) {
	return nil, nil
}

func (s *SamFSServer) Create(ctx context.Context, req *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
	return nil, nil
}

func (s *SamFSServer) Remove(ctx context.Context, req *pb.LocalDirectoryRequest) (*pb.StatusReply, error) {
	return nil, nil
}

func (s *SamFSServer) Mkdir(ctx context.Context, req *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
	return nil, nil
}

func (s *SamFSServer) Rmdir(ctx context.Context, req *pb.LocalDirectoryRequest) (*pb.StatusReply, error) {
	return nil, nil
}
