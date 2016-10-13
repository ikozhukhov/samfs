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

	return &SamFSClient{
		samFS:      samFS,
		fuseServer: fuseServer,
	}, nil
}

func (c *SamFSClient) Run() {
	c.fuseServer.Serve()
}
