package samfs

type SamFSClient struct{}

func NewClient() (*SamFSClient, error) {
  c := &SamFSClient{}
  return c, nil
}
