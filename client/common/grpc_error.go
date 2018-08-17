package common

import (
	"fmt"

	"github.com/samoslab/nebula/client/errcode"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StatusErrFromError return grpc status error
func StatusErrFromError(err error) (int, string) {
	st, _ := status.FromError(err)
	return int(st.Code()), st.Message()
}

func NewStatusErr(code uint32, errmsg string) error {
	return status.New(codes.Code(code), errmsg).Err()
}

func NewStatus(code errcode.Status, err error) error {
	return status.New(codes.Code(code), fmt.Sprintf("%s:%v", code.String(), err)).Err()
}
