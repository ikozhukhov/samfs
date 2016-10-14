package samfs

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type SamFSClient struct {
	samFS      *SamFs
	fuseServer *fuse.Server
}

func NewClient(server, port, mountDir *string) (*SamFSClient, error) {
	samFS, fsErr := NewSamFs(&SamFsOptions{
		server: *server,
		port:   *port,
	})
	if fsErr != nil {
		return nil, fsErr
	}
	pOpts := pathfs.PathNodeFsOptions{
		Debug: true,
	}
	pathFs := pathfs.NewPathNodeFs(samFS, &pOpts)
	opts := nodefs.Options{
		Debug: true,
	}
	connector := nodefs.NewFileSystemConnector(pathFs.Root(), &opts)

	mountOpts := &fuse.MountOptions{
		AllowOther: true,
		Name:       "samfs://" + *server + ":" + *port,
	}
	fuseServer, err := fuse.NewServer(connector.RawFS(), *mountDir, mountOpts)
	if err != nil {
		return nil, err
	}

	return &SamFSClient{
		samFS:      samFS,
		fuseServer: fuseServer,
	}, nil
}

func (c *SamFSClient) Run() {
	c.fuseServer.Serve()
}
