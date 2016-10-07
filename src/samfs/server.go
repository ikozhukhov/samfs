package samfs

import (
  pb "github.com/smihir/samfs/src/proto"
  "golang.org/x/net/context"
)

type SamFSServer struct{}

var _ pb.NFSServer = SamFSServer{} // or &myType{} or [&]myType if scalar

func NewServer() (*SamFSServer, error) {
  s := &SamFSServer{}
  return s, nil
}

func (s SamFSServer) Mount(context.Context, *pb.MountRequest) (*pb.FileHandleReply, error) {
  return nil, nil
}

func (s SamFSServer) Lookup(context.Context, *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
  return nil, nil
}

func (s SamFSServer) Read(context.Context, *pb.ReadRequest) (*pb.ReadReply, error) {
  return nil, nil
}

func (s SamFSServer) Write(context.Context, *pb.WriteRequest) (*pb.StatusReply, error) {
  return nil, nil
}

func (s SamFSServer) Commit(context.Context, *pb.CommitRequest) (*pb.StatusReply, error) {
  return nil, nil
}

func (s SamFSServer) Create(context.Context, *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
  return nil, nil
}

func (s SamFSServer) Remove(context.Context, *pb.LocalDirectoryRequest) (*pb.StatusReply, error) {
  return nil, nil
}

func (s SamFSServer) Mkdir(context.Context, *pb.LocalDirectoryRequest) (*pb.FileHandleReply, error) {
  return nil, nil
}

func (s SamFSServer) Rmdir(context.Context, *pb.LocalDirectoryRequest) (*pb.StatusReply, error) {
  return nil, nil
}
