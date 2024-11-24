package catalystgo

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
)

func newListener(port int) (net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, errors.Errorf("listen to port %d: %v", port, err)
	}
	return lis, nil
}
