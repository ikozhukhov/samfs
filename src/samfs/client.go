package samfs

import (
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	pb "github.com/smihir/samfs/src/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type SamFSClient struct {
	server     string
	port       string
	nfsClient  pb.NFSClient
	clientConn *grpc.ClientConn
	fuseServer *fuse.Server
}

func NewClient(server, port, mountDir *string) (*SamFSClient, error) {
	c := &SamFSClient{}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := grpc.DialContext(ctx, *server+":"+*port,
		grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c.server = *server
	c.port = *port
	c.nfsClient = pb.NewNFSClient(conn)
	c.clientConn = conn
	samFS := NewSamFs(&SamFsOptions{})
	pathFs := pathfs.NewPathNodeFs(samFS, nil)
	connector := nodefs.NewFileSystemConnector(pathFs.Root(), nil)
	mountOpts := &fuse.MountOptions{
		AllowOther: true,
		Name:       "samfs:",
	}
	fuseServer, err := fuse.NewServer(connector.RawFS(), *mountDir, mountOpts)
	if err != nil {
		return nil, err
	}
	c.fuseServer = fuseServer
	return c, nil
}

func (c *SamFSClient) Run() {
	c.fuseServer.Serve()
}
