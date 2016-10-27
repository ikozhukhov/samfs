package samfs

import (
	"errors"
	"fmt"
	"io"
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
	defaultPermission os.FileMode = 0766
)

type serverInfo struct {
	commitCount  uint64
	createCount  uint64
	lookupCount  uint64
	mkdirCount   uint64
	mountCount   uint64
	readCount    uint64
	readDirCount uint64
	removeCount  uint64
	renameCount  uint64
	writeCount   uint64
	getAttrCount uint64
	rmDirCount   uint64
}

type SamFSServer struct {
	rootDirectory  string
	rootFileHandle *pb.FileHandle

	port       string
	grpcServer *grpc.Server

	//sessionID is randomly generated every time server starts;
	//it is used to detect server crashes
	sessionID int64

	info *serverInfo
	tick *time.Ticker
}

var _ pb.NFSServer = &SamFSServer{}

func NewServer(rootDirectory string, port string) (*SamFSServer, error) {
	inum, gnum, err := GetInodeAndGenerationNumbers(rootDirectory)
	if err != nil {
		glog.Errorf("failed to get inode and generation number for root "+
			"directory :: %v", err)
		return nil, err
	}

	rootFileHandle := &pb.FileHandle{
		Path:             "/",
		InodeNumber:      inum,
		GenerationNumber: gnum,
	}

	s := &SamFSServer{
		rootDirectory:  rootDirectory,
		rootFileHandle: rootFileHandle,
		// TODO(mihir): make port number configurable
		port: ":" + port,
		tick: time.NewTicker(10 * time.Second),
		info: &serverInfo{},
	}
	go func() {
		for _ = range s.tick.C {
			glog.Infof("%+v", s.info)
		}
	}()

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
	return gs.Serve(lis)
}

func (s *SamFSServer) Stop() error {
	s.grpcServer.GracefulStop()
	s.tick.Stop()
	return nil
}

func (s *SamFSServer) Mount(ctx context.Context,
	req *pb.MountRequest) (*pb.FileHandleReply, error) {
	glog.V(3).Info("recevied mount request")

	resp := &pb.FileHandleReply{
		FileHandle: s.rootFileHandle,
	}

	return resp, nil
}

func (s *SamFSServer) Lookup(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
	glog.V(3).Infof(`received lookup request for "%s"`, req.Name)
	s.info.lookupCount++

	//validate incoming directory file handle
	err := s.verifyFileHandle(req.DirectoryFileHandle)
	if err != nil {
		glog.Errorf(err.Error())
		return nil, err
	}

	//get info about file being looked up
	directoryPath := path.Join(s.rootDirectory, req.DirectoryFileHandle.Path)
	filePath := path.Join(directoryPath, req.Name)
	inum, gnum, err := GetInodeAndGenerationNumbers(filePath)
	if err != nil {
		glog.V(3).Infof("failed to get inode and generation number for %s :: %v\n",
			filePath, err)
		return nil, err
	}

	fsFilePath := path.Join(req.DirectoryFileHandle.Path, req.Name)
	fileHandle := &pb.FileHandle{
		Path:             fsFilePath,
		InodeNumber:      inum,
		GenerationNumber: gnum,
	}

	resp := &pb.FileHandleReply{
		FileHandle: fileHandle,
	}

	return resp, nil
}

func (s *SamFSServer) GetAttr(ctx context.Context,
	req *pb.FileHandleRequest) (*pb.GetAttrReply, error) {
	glog.V(3).Infof(`received GetAttr request for "%s"`, req.FileHandle.Path)
	s.info.getAttrCount++

	//validate incoming file handle
	err := s.verifyFileHandle(req.FileHandle)
	if err != nil {
		glog.Errorf(err.Error())
		return nil, err
	}

	filePath := path.Join(s.rootDirectory, req.FileHandle.Path)
	fd, oErr := syscall.Open(filePath, syscall.O_RDONLY, 0)
	defer func() {
		e := syscall.Close(fd)
		if e != nil {
			glog.Errorf("unable to close file %s :: %s", filePath,
				e.Error())
		}
	}()

	if oErr != nil {
		glog.Errorf("could not get open file %s :: %v", filePath, oErr)
		return nil, oErr
	}
	var stat syscall.Stat_t
	fsErr := syscall.Fstat(fd, &stat)
	if fsErr != nil {
		glog.Errorf("could not get stat on file %s :: %v", filePath, fsErr)
		return nil, err
	}

	attr := StatToProtoAttr(&stat)

	return attr, nil
}

func (s *SamFSServer) Readdir(ctx context.Context,
	req *pb.FileHandleRequest) (*pb.ReaddirReply, error) {
	glog.V(3).Infof("received Readdir request root: %s, path: %s", s.rootDirectory,
		req.FileHandle.Path)
	s.info.readDirCount++

	//validate incoming file handle
	err := s.verifyFileHandle(req.FileHandle)
	if err != nil {
		glog.Errorf(err.Error())
		return nil, err
	}

	filePath := path.Join(s.rootDirectory, req.FileHandle.Path)
	fd, err := os.Open(filePath)
	if err != nil {
		glog.Errorf("could not get open file %s :: %v", filePath, err)
		return nil, err
	}
	defer fd.Close()

	entries, err := fd.Readdir(0)
	if err != nil {
		glog.Errorf("could not readdir file %s :: %v", filePath, err)
		return nil, err
	}

	respEntries := make([]*pb.DirEntry, len(entries), cap(entries))
	for i, entry := range entries {
		respEntries[i] = &pb.DirEntry{
			Name: entry.Name(),
			Mode: uint32(entry.Mode()),
		}
	}

	resp := &pb.ReaddirReply{
		Entries: respEntries,
	}

	return resp, nil
}

func (s *SamFSServer) Read(ctx context.Context,
	req *pb.ReadRequest) (*pb.ReadReply, error) {
	glog.V(3).Info("received read request")
	s.info.readCount++

	//validate incoming file handle
	err := s.verifyFileHandle(req.FileHandle)
	if err != nil {
		glog.Errorf(err.Error())
		return nil, err
	}

	filePath := path.Join(s.rootDirectory, req.FileHandle.Path)
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
	if err != nil && err != io.EOF {
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
	glog.V(3).Info("recevied write request")
	s.info.writeCount++

	//validate incoming file handle
	err := s.verifyFileHandle(req.FileHandle)
	if err != nil {
		glog.Errorf(err.Error())
		return nil, err
	}

	filePath := path.Join(s.rootDirectory, req.FileHandle.Path)
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

	if req.ShouldCommit {
		glog.V(3).Infof("syncing file %s write in write()", req.FileHandle.Path)
		err = fd.Sync()
		if err != nil {
			glog.Errorf("could not perform fsync on file %s :: %v\n",
				req.FileHandle.Path, err)
			return nil, err
		}
	}

	resp := &pb.StatusReply{
		Success:         true,
		ServerSessionID: s.sessionID,
	}

	return resp, nil
}

func (s *SamFSServer) Commit(ctx context.Context,
	req *pb.CommitRequest) (*pb.StatusReply, error) {
	glog.V(3).Info("recevied commit request")
	s.info.commitCount++

	//validate incoming file handle
	err := s.verifyFileHandle(req.FileHandle)
	if err != nil {
		glog.Errorf(err.Error())
		return nil, err
	}

	filePath := path.Join(s.rootDirectory, req.FileHandle.Path)
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
	glog.V(3).Infof("recevied create request root: %s, path: %s", s.rootDirectory,
		req.DirectoryFileHandle.Path)
	s.info.createCount++

	//validate incoming directory file handle
	err := s.verifyFileHandle(req.DirectoryFileHandle)
	if err != nil {
		glog.Errorf(err.Error())
		return nil, err
	}

	directoryPath := path.Join(s.rootDirectory, req.DirectoryFileHandle.Path)
	filePath := path.Join(directoryPath, req.Name)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0766)
	if err != nil {
		glog.Errorf("Failed to create file at path %s :: %v\n", filePath, err)
		return nil, err
	}
	file.Close()

	err = flush(directoryPath)
	if err != nil {
		glog.Warningf("failed to flush parent directory on Create :: %v\n", err)
	}

	inum, gnum, err := GetInodeAndGenerationNumbers(filePath)
	if err != nil {
		glog.Errorf("failed to get inode and generation number for %s :: %v\n",
			filePath, err)
		err = os.Remove(filePath)
		if err != nil {
			glog.Errorf("failed to remove file after not getting its info :: %v\n",
				err)
		}
		return nil, err
	}

	fsFilePath := path.Join(req.DirectoryFileHandle.Path, req.Name)
	fileHandle := &pb.FileHandle{
		Path:             fsFilePath,
		InodeNumber:      inum,
		GenerationNumber: gnum,
	}

	resp := &pb.FileHandleReply{
		FileHandle: fileHandle,
	}

	return resp, nil
}

func (s *SamFSServer) Remove(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.StatusReply, error) {
	glog.V(3).Info("recevied remove request")
	s.info.removeCount++
	return s.remove(ctx, req)
}

func (s *SamFSServer) Mkdir(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
	glog.V(3).Info("recevied mkdir request")
	s.info.mkdirCount++

	//validate incoming directory file handle
	err := s.verifyFileHandle(req.DirectoryFileHandle)
	if err != nil {
		glog.Errorf(err.Error())
		return nil, err
	}

	directoryPath := path.Join(s.rootDirectory, req.DirectoryFileHandle.Path)
	filePath := path.Join(directoryPath, req.Name)
	err = os.Mkdir(filePath, defaultPermission)
	if err != nil {
		glog.Errorf("Failed to make directory at path %s :: %v\n", filePath, err)
		return nil, err
	}

	err = flush(directoryPath)
	if err != nil {
		glog.Warningf("failed to flush parent directory on Rmdir :: %v\n", err)
	}

	inum, gnum, err := GetInodeAndGenerationNumbers(filePath)
	if err != nil {
		glog.Errorf("failed to get inode and generation number for %s :: %v\n",
			filePath, err)
		err = os.Remove(filePath)
		if err != nil {
			glog.Errorf("failed to remove file after not getting its info :: %v\n",
				err)
		}
		return nil, err
	}

	fsFilePath := path.Join(req.DirectoryFileHandle.Path, req.Name)
	fileHandle := &pb.FileHandle{
		Path:             fsFilePath,
		InodeNumber:      inum,
		GenerationNumber: gnum,
	}

	resp := &pb.FileHandleReply{
		FileHandle: fileHandle,
	}

	return resp, nil
}

func (s *SamFSServer) Rmdir(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.StatusReply, error) {
	glog.V(3).Info("recevied Rmdir request")
	s.info.rmDirCount++
	return s.remove(ctx, req)
}

func (s *SamFSServer) Rename(ctx context.Context,
	req *pb.RenameRequest) (*pb.StatusReply, error) {
	glog.V(3).Info("received Rename request from %s to %s", req.FromName, req.ToName)
	s.info.renameCount++
	//validating incoming directory file handle
	fromErr := s.verifyFileHandle(req.FromDirHandle)
	if fromErr != nil {
		glog.Errorf(fromErr.Error())
		return nil, fromErr
	}

	toErr := s.verifyFileHandle(req.ToDirHandle)
	if toErr != nil {
		glog.Errorf(toErr.Error())
		return nil, toErr
	}

	fromDirPath := path.Join(s.rootDirectory, req.FromDirHandle.Path)
	fromFilePath := path.Join(fromDirPath, req.FromName)
	toDirPath := path.Join(s.rootDirectory, req.ToDirHandle.Path)
	toFilePath := path.Join(toDirPath, req.ToName)

	renErr := os.Rename(fromFilePath, toFilePath)
	if renErr != nil {
		glog.Errorf(renErr.Error())
		return nil, renErr
	}

	err := flush(toFilePath)
	if err != nil {
		glog.Warningf("failed to flush file on rename :: %v", err)
	}
	resp := &pb.StatusReply{
		Success: true,
	}

	return resp, nil
}

//common methods

func (s *SamFSServer) remove(ctx context.Context,
	req *pb.LocalDirectoryRequest) (*pb.StatusReply, error) {
	//validate incoming directory file handle
	err := s.verifyFileHandle(req.DirectoryFileHandle)
	if err != nil {
		glog.Errorf(err.Error())
		return nil, err
	}

	directoryPath := path.Join(s.rootDirectory, req.DirectoryFileHandle.Path)
	filePath := path.Join(directoryPath, req.Name)
	err = os.Remove(filePath)
	if err != nil {
		glog.Errorf("Failed to remove file/directory at path %s :: %v\n", filePath,
			err)
		return nil, err
	}

	err = flush(directoryPath)
	if err != nil {
		glog.Warningf("failed to flush parent directory on remove :: %v\n", err)
	}

	resp := &pb.StatusReply{
		Success: true,
	}

	return resp, nil
}

func (s *SamFSServer) verifyFileHandle(fileHandle *pb.FileHandle) error {
	filePath := path.Join(s.rootDirectory, fileHandle.Path)
	inum, gnum, err := GetInodeAndGenerationNumbers(filePath)
	if err != nil {
		glog.Errorf("failed to get inode and generation number for %s :: %v\n",
			fileHandle.Path, err)
		return err
	}

	if inum != fileHandle.InodeNumber || gnum != fileHandle.GenerationNumber {
		errStr := fmt.Sprintf("file handle for %s is not valid\n", fileHandle.Path)
		glog.Errorf(errStr)
		return errors.New(errStr)
	}

	return nil
}

func flush(path string) error {
	fd, err := os.Open(path)
	if err != nil {
		glog.Errorf("failed to open file/dir at path %s :: %v\n", path, err)
		return err
	}
	defer fd.Close()

	err = fd.Sync()
	if err != nil {
		glog.Errorf("could not fsync file/dir at path %s :: %v\n", path, err)
		return err
	}

	return nil
}
