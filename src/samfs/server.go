package samfs

import (
	"net"
	"os"
	"path"

	"github.com/boltdb/bolt"
	"github.com/golang/glog"
	pb "github.com/smihir/samfs/src/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const dbFileName string = "samfs.db"

type SamFSServer struct {
	rootDirectory string
	db            *bolt.DB
	port          string
	grpcServer    *grpc.Server
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
		// TODO(mihir): make port number configurable
		port: ":24100",
	}

	return s, nil
}

//TODO (arman): run() and stop() where stop closes database
func (s *SamFSServer) Run() error {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		glog.Fatalf("falied to listen on port :: %s(err=%s)", s.port, err.Error())
		return err
	}

	gs := grpc.NewServer()
	pb.RegisterNFSServer(gs, s)
	s.grpcServer = gs
	gs.Serve(lis)

	return nil
}

func (s *SamFSServer) Stop() error {
	s.grpcServer.GracefulStop()
	return nil
}

func (s *SamFSServer) Mount(ctx context.Context, req *pb.MountRequest) (*pb.FileHandleReply, error) {
	glog.Info("recevied mount request")
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
	directoryPath := path.Join(s.rootDirectory, req.DirectoryFileHandle.Path)
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		glog.Errorf("path %s does not exist :: %v\n", req.DirectoryFileHandle.Path, err)
		return nil, err
	}

	//TODO (arman): check version number of local directory against db

	fsFilePath := path.Join(req.DirectoryFileHandle.Path, req.Name)
	//TODO (arman): add fsFilePath and its *new* version number to db

	filePath := path.Join(directoryPath, req.Name)
	file, err := os.Create(filePath)
	if err != nil {
		glog.Errorf("Failed to create file at path %s :: %v\n", filePath, err)
		return nil, err
	}
	file.Close()

	fileHandle := &pb.FileHandle{
		Path:    fsFilePath,
		Version: 0, //TODO (arman): return the new version number
	}

	resp := &pb.FileHandleReply{
		FileHandle: fileHandle,
	}

	return resp, nil
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
