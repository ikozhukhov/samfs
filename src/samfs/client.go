package samfs

import (
	"time"

	pb "github.com/smihir/samfs/src/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type SamFSClient struct {
	server     string
	port       string
	nfsClient  pb.NFSClient
	clientConn *grpc.ClientConn
}

func NewClient(server *string, port *string) (*SamFSClient, error) {
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
	return c, nil
}
