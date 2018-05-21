package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/samoslab/nebula/client/common"
)

func GetAllPackageHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !s.CanBeWork() {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("register first"))
			return
		}
		log := s.cm.Log
		w.Header().Set("Accept", "application/json")

		if !validMethod(ctx, w, r, []string{http.MethodGet}) {
			return
		}

		result, err := s.cm.OM.GetAllPackages()
		code := 0
		errmsg := ""
		if err != nil {
			log.Errorf("get all packages error %v", err)
			code = 1
			errmsg = err.Error()
			result = nil
		}

		rsp, err := common.MakeUnifiedHTTPResponse(code, result, errmsg)
		if err != nil {
			errorResponse(ctx, w, http.StatusBadRequest, err)
			return
		}
		if err := JSONResponse(w, rsp); err != nil {
			fmt.Printf("error %v\n", err)
		}
	}
}
