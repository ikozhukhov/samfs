package samfs

import (
	"os"
	"path"

	"github.com/boltdb/bolt"
	"github.com/golang/glog"
	pb "github.com/smihir/samfs/src/proto"
	"golang.org/x/net/context"
)

const dbFileName string = "samfs.db"

type SamFSServer struct {
	rootDirectory string
	db            *bolt.DB
}

var _ pb.NFSServer = &SamFSServer{}

func NewServer(rootDirectory string) (*SamFSServer, error) {
	if _, err := os.Stat(rootDirectory); os.IsNotExist(err) {
		glog.Errorf("path %s does not exist :: %v\n", rootDirectory, err)
		return nil, err
	}

	glog.Infof("FS root = %s\n", rootDirectory)

	dbPath := path.Join(path.Dir(rootDirectory), dbFileName)
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		glog.Errorf("failed to open inode database :: %v", err)
		return nil, err
	}

	s := &SamFSServer{
		rootDirectory: rootDirectory,
		db:            db,
	}

	return s, nil
}

//TODO (arman): run() and stop() where stop closes database

func (s *SamFSServer) Mount(ctx context.Context, req *pb.MountRequest) (*pb.FileHandleReply, error) {
	fileHandle := &pb.FileHandle{
		Path:    "/",
		Version: 0,
	}

	resp := &pb.FileHandleReply{
		FileHandle: fileHandle,
	}

	return resp, nil
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
