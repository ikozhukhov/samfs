package samfs

type samfsServer struct{}

func NewServer() (*samfsServer, error) {
	s := &samfsServer{}
	return s, nil
}
