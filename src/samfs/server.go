package samfs

import (
	"errors"
	"math/rand"
	"net"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/golang/glog"
	pb "github.com/smihir/samfs/src/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	dbFileName        string      = "samfs.db"
	defaultPermission os.FileMode = 0666
)

type SamFSServer struct {
	rootDirectory string
	db            *DB
	port          string
	grpcServer    *grpc.Server

	//sessionID is randomly generated every time server starts;
	//it is used to detect server crashes
	sessionID int64
}

var _ pb.NFSServer = &SamFSServer{}

func NewServer(rootDirectory string) (*SamFSServer, error) {
	if _, err := os.Stat(rootDirectory); os.IsNotExist(err) {
		glog.Errorf("path %s does not exist :: %v\n", rootDirectory, err)
		return nil, err
	}

	glog.Infof("FS root = %s\n", rootDirectory)

	dbPath := path.Join(path.Dir(rootDirectory), dbFileName)
	db, err := NewDB(dbPath)
	if err != nil {
		glog.Errorf("failed to create new instance of database :: %v", err)
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

	rand.Seed(time.Now().UnixNano())
	s.sessionID = rand.Int63()
	glog.Infof("starting new server with sessionID %d", s.sessionID)

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

func (s *SamFSServer) Mount(ctx context.Context,
	req *pb.MountRequest) (*pb.FileHandleReply, error) {
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

func (s *SamFSServer) Lookup(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
	glog.Info("received lookup request")

	directoryPath := path.Join(s.rootDirectory, req.DirectoryFileHandle.Path)
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		glog.Errorf("path %s does not exist :: %v\n", req.DirectoryFileHandle.Path,
			err)
		return nil, err
	}

	//TODO (arman): check version number of local directory against db

	fsFilePath := path.Join(req.DirectoryFileHandle.Path, req.Name)
	//TODO (arman): add fsFilePath and its *new* version number to db

	filePath := path.Join(directoryPath, req.Name)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		glog.Errorf("requested file %s does not exist :: %v\n", filePath, err)
		return nil, err
	}

	fileHandle := &pb.FileHandle{
		Path:    fsFilePath,
		Version: 0, //TODO (arman): return file's version number
	}

	resp := &pb.FileHandleReply{
		FileHandle: fileHandle,
	}

	return resp, nil
}

func (s *SamFSServer) Read(ctx context.Context,
	req *pb.ReadRequest) (*pb.ReadReply, error) {
	glog.Info("received read request")

	filePath := path.Join(s.rootDirectory, req.FileHandle.Path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		glog.Errorf("file %s does not exist :: %v\n", req.FileHandle.Path,
			err)
		return nil, err
	}

	//TODO (arman): check version number of the file against db

	fd, err := os.Open(filePath)
	if err != nil {
		glog.Errorf("could not open file %s :: %v\n", req.FileHandle.Path, err)
		return nil, err
	}
	defer fd.Close()

	data := make([]byte, req.Size, req.Size)
	if data == nil {
		errStr := "couldn't allocate memory"
		glog.Errorf(errStr)
		return nil, errors.New(errStr)
	}

	n, err := fd.ReadAt(data, req.Offset)
	if err != nil {
		glog.Errorf("failed to read file %s :: %v\n", req.FileHandle.Path, err)
		return nil, err
	}

	resp := &pb.ReadReply{
		Data: data[:n],
		Size: int64(n),
	}

	return resp, nil
}

func (s *SamFSServer) Write(ctx context.Context,
	req *pb.WriteRequest) (*pb.StatusReply, error) {
	glog.Info("recevied write request")

	filePath := path.Join(s.rootDirectory, req.FileHandle.Path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		glog.Errorf("file %s does not exist :: %v\n", req.FileHandle.Path,
			err)
		return nil, err
	}

	//TODO (arman): check version number of the file against db

	fd, err := os.OpenFile(filePath, os.O_WRONLY, defaultPermission)
	if err != nil {
		glog.Errorf("could not open file %s :: %v\n", req.FileHandle.Path, err)
		return nil, err
	}
	defer fd.Close()

	_, err = fd.WriteAt(req.Data[:req.Size], req.Offset)
	if err != nil {
		glog.Errorf("failed to write file %s :: %v\n", req.FileHandle.Path, err)
		return nil, err
	}

	resp := &pb.StatusReply{
		Success:         true,
		ServerSessionID: s.sessionID,
	}

	return resp, nil
}

func (s *SamFSServer) Commit(ctx context.Context,
	req *pb.CommitRequest) (*pb.StatusReply, error) {
	glog.Info("recevied commit request")

	filePath := path.Join(s.rootDirectory, req.FileHandle.Path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		glog.Errorf("file %s does not exist :: %v\n", req.FileHandle.Path,
			err)
		return nil, err
	}

	//TODO (arman): check version number of the file against db

	fd, err := os.OpenFile(filePath, os.O_WRONLY, defaultPermission)
	if err != nil {
		glog.Errorf("could not open file %s :: %v\n", req.FileHandle.Path, err)
		return nil, err
	}
	defer fd.Close()

	err = fd.Sync()
	if err != nil {
		glog.Errorf("could not perform fsync on file %s :: %v\n",
			req.FileHandle.Path, err)
		return nil, err
	}

	resp := &pb.StatusReply{
		Success:         true,
		ServerSessionID: s.sessionID,
	}

	return resp, nil
}

func (s *SamFSServer) Create(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
	glog.Info("recevied create request")

	directoryPath := path.Join(s.rootDirectory, req.DirectoryFileHandle.Path)
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		glog.Errorf("path %s does not exist :: %v\n", req.DirectoryFileHandle.Path,
			err)
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

func (s *SamFSServer) Remove(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.StatusReply, error) {
	glog.Info("recevied remove request")
	return s.remove(ctx, req)
}

func (s *SamFSServer) Mkdir(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
	glog.Info("recevied mkdir request")

	directoryPath := path.Join(s.rootDirectory, req.DirectoryFileHandle.Path)
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		glog.Errorf("path %s does not exist :: %v\n", req.DirectoryFileHandle.Path,
			err)
		return nil, err
	}

	//TODO (arman): check version number of local directory against db

	fsFilePath := path.Join(req.DirectoryFileHandle.Path, req.Name)
	//TODO (arman): add fsFilePath and its *new* version number to db

	filePath := path.Join(directoryPath, req.Name)
	err := os.Mkdir(filePath, defaultPermission)
	if err != nil {
		glog.Errorf("Failed to make directory at path %s :: %v\n", filePath, err)
		return nil, err
	}

	fileHandle := &pb.FileHandle{
		Path:    fsFilePath,
		Version: 0, //TODO (arman): return the new version number
	}

	resp := &pb.FileHandleReply{
		FileHandle: fileHandle,
	}

	return resp, nil
}

func (s *SamFSServer) Rmdir(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.StatusReply, error) {
	glog.Info("recevied rmdir request")
	return s.remove(ctx, req)
}

//common methods

func (s *SamFSServer) remove(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.StatusReply, error) {
	directoryPath := path.Join(s.rootDirectory, req.DirectoryFileHandle.Path)
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		glog.Errorf("path %s does not exist :: %v\n", req.DirectoryFileHandle.Path,
			err)
		return nil, err
	}

	//TODO (arman): check version number of local directory against db

	//since we first check whether file/directory exists
	//before using it, we do not need to
	//update its version number on delete.

	filePath := path.Join(directoryPath, req.Name)
	err := os.Remove(filePath)
	if err != nil {
		glog.Errorf("Failed to remove file/directory at path %s :: %v\n", filePath,
			err)
		return nil, err
	}

	resp := &pb.StatusReply{
		Success: true,
	}

	return resp, nil
}

func (s *SamFSServer) GetInodeAndGenerationNumbers(filePath string) (uint64, int, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(filePath, &stat); err != nil {
		return 0, 0, err
	}

	genNumber, err := getGenerationNumber(filePath)
	if err != nil {
		return 0, 0, err
	}

	return stat.Ino, genNumber, nil
}
