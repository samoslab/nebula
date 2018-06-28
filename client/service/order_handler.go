package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/samoslab/nebula/client/common"
)

// GetAllPackageHandler all package handler
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

// GetPackageInfoHandler get package handler
func GetPackageInfoHandler(s *HTTPServer) http.HandlerFunc {
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

		id := r.URL.Query().Get("id")
		if id == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("need paras id"))
			return
		}
		result, err := s.cm.OM.GetPackageInfo(id)
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

// BuyPackageReq request struct for buy package
type BuyPackageReq struct {
	ID       string `json:"id"`
	Canceled bool   `json:"canceled"`
	Quanlity uint32 `json:"quanlity"`
}

// BuyPackageHandler buy package handler
func BuyPackageHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !s.CanBeWork() {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("register first"))
			return
		}
		log := s.cm.Log
		w.Header().Set("Accept", "application/json")

		if !validMethod(ctx, w, r, []string{http.MethodPost}) {
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			errorResponse(ctx, w, http.StatusUnsupportedMediaType, errors.New("Invalid content type"))
			return
		}

		req := &BuyPackageReq{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&req); err != nil {
			err = fmt.Errorf("Invalid json request body: %v", err)
			errorResponse(ctx, w, http.StatusBadRequest, err)
			return
		}

		defer r.Body.Close()
		if req.ID == "" || req.Quanlity == 0 {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("argument id or quanlity must not empty"))
			return
		}

		result, err := s.cm.OM.BuyPackage(req.ID, req.Canceled, req.Quanlity)
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

// DiscountPackageReq request struct for discount
type DiscountPackageReq struct {
	ID string `json:"id"`
}

// DiscountPackageHandler discount handler
func DiscountPackageHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !s.CanBeWork() {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("register first"))
			return
		}
		log := s.cm.Log
		w.Header().Set("Accept", "application/json")
		defer r.Body.Close()

		if !validMethod(ctx, w, r, []string{http.MethodGet}) {
			return
		}

		id := r.URL.Query().Get("id")
		if id == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("id must not empty"))
			return
		}

		fmt.Printf("package %d\n", id)

		result, err := s.cm.OM.DiscountPackage(id)
		code := 0
		errmsg := ""
		if err != nil {
			log.Errorf("discount package error %v", err)
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

// MyAllOrderHandler return all orders
func MyAllOrderHandler(s *HTTPServer) http.HandlerFunc {
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

		expired := r.URL.Query().Get("expired")
		boolExpired := true
		if expired != "" {
			var err error
			boolExpired, err = strconv.ParseBool(expired)
			if err != nil {
				errorResponse(ctx, w, http.StatusBadRequest, errors.New("expired args wrong, need bool"))
				return
			}
		}
		result, err := s.cm.OM.MyAllOrders(boolExpired)
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

// GetOrderInfoHandler returns order
func GetOrderInfoHandler(s *HTTPServer) http.HandlerFunc {
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

		id := r.URL.Query().Get("orderid")
		if id == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("need paras orderid"))
			return
		}
		result, err := s.cm.OM.GetOrderInfo(id)
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

// PayOrdorReq request struct for pay order
type PayOrderReq struct {
	ID string `json:"order_id"`
}

// PayOrderHandler  pay order handler
func PayOrderHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !s.CanBeWork() {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("register first"))
			return
		}
		log := s.cm.Log
		w.Header().Set("Accept", "application/json")

		if !validMethod(ctx, w, r, []string{http.MethodPost}) {
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			errorResponse(ctx, w, http.StatusUnsupportedMediaType, errors.New("Invalid content type"))
			return
		}

		req := &PayOrderReq{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&req); err != nil {
			err = fmt.Errorf("Invalid json request body: %v", err)
			errorResponse(ctx, w, http.StatusBadRequest, err)
			return
		}

		defer r.Body.Close()
		if req.ID == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("argument order_id must not empty"))
			return
		}

		result, err := s.cm.OM.PayOrdor(req.ID)
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

// RechargeAddressHandler get balance handler
func RechargeAddressHandler(s *HTTPServer) http.HandlerFunc {
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

		result, err := s.cm.OM.RechargeAddress()
		code := 0
		errmsg := ""
		if err != nil {
			log.Errorf("get usage amount error %v", err)
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

// UsageAmountHandler get usage amount info
func UsageAmountHandler(s *HTTPServer) http.HandlerFunc {
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

		result, err := s.cm.OM.UsageAmount()
		code := 0
		errmsg := ""
		if err != nil {
			log.Errorf("get usage amount error %v", err)
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
