package common

import (
	"fmt"

	"google.golang.org/grpc/status"
)

func StatusErrFromError(err error) error {
	st, _ := status.FromError(err)
	return fmt.Errorf("code %d, msg %s", st.Code(), st.Message())
}
