package samfs

type samfsClient struct{}

func NewClient() (*samfsClient, error) {
	c := &samfsClient{}
	return c, nil
}
