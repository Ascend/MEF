// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package limiter implement a token bucket limiter
package limiter

import (
	"context"
	"errors"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"huawei.com/mindx/common/cache"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
)

const (
	kilo = 1000.0
	// DefaultDataLimit  default http body limit size
	DefaultDataLimit      = 1024 * 1024 * 10
	defaultMaxConcurrency = 1024
	second5               = 5
	maxStringLen          = 20
	// DefaultCacheSize  default cache size
	DefaultCacheSize = 1024 * 100
	arrLen           = 2
	// IPReqLimitReg  ip request limit regex string
	IPReqLimitReg = "^[1-9]\\d{0,2}/[1-9]\\d{0,2}$"
)

type limitHandler struct {
	limitBytes        int64
	ipLimiterFactory  IndependentLimiterFactory
	ipLimiters        *cache.ConcurrencyLRUCache
	conCurrentHandler concurrentHandler
	log               bool
}

type concurrentHandler interface {
	handle(w http.ResponseWriter, r *http.Request, ctx context.Context)
}

type concHandlerWithCtxAwareLimiter struct {
	limiter ContextAwareLimiter
	handler http.Handler
	log     bool
}

// HandlerConfig the configuration of the limitHandler
type HandlerConfig struct {
	// PrintLog whether you need print access log, when use gin framework, suggest to set false,otherwise set true
	PrintLog bool
	// Method only allow setting  http method pass
	Method string
	// LimitBytes set the max http body size
	LimitBytes int64
	// TotalConCurrency set the program total concurrent http request
	TotalConCurrency int
	// IPConCurrency set the signle IP concurrent http request "2/1sec"
	IPConCurrency string
	// CacheSize the local cacheSize
	CacheSize int
}

// HandlerConfigV3 is an extent configuration of the limitHandler, allow to ip-limiter enable burst requests
type HandlerConfigV3 struct {
	HandlerConfig
	// IPBurst define token capacity, which allow single IP limiter enable concurrency
	IPBurst int
}

// StatusResponseWriter the writer record the http status
type StatusResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier
	Status int
}

// WriteHeader override the WriteHeader method
func (w *StatusResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.Status = status
}

// Pusher implements http.Pusher
func (w *StatusResponseWriter) Pusher() http.Pusher {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}

// ServeHTTP implement http.Handler
func (h *limitHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, h.limitBytes)
	ctx := initContext(req)
	path := req.URL.Path
	clientUserAgent := req.UserAgent()
	clientIP := utils.ClientIP(req)
	if clientIP != "" && !h.limitSingleIp(clientIP) {
		if h.log {
			hwlog.RunLog.WarnfWithCtx(ctx, "Total reject request:%s: %s <%3d> |%15s |%s |%d ", req.Method, path,
				http.StatusTooManyRequests, clientIP, clientUserAgent, syscall.Getuid())
		}
		http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
		return
	}
	h.conCurrentHandler.handle(w, req, ctx)
}

func (hcl *concHandlerWithCtxAwareLimiter) handle(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	path := r.URL.Path
	clientUserAgent := r.UserAgent()
	clientIP := utils.ClientIP(r)
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	start := time.Now()
	if !hcl.limiter.Allow(cancelCtx) {
		cancelFunc()
		if hcl.log {
			hwlog.RunLog.WarnfWithCtx(ctx, "Total reject request:%s: %s <%3d> |%15s |%s |%d ", r.Method,
				path, http.StatusTooManyRequests, clientIP, clientUserAgent, syscall.Getuid())
		}
		http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
		return
	}
	statusRes := newResponse(w)
	hcl.handler.ServeHTTP(statusRes, r)
	stop := time.Since(start)
	cancelFunc()
	latency := int(math.Ceil(float64(stop.Nanoseconds()) / kilo / kilo))
	if hcl.log {
		hwlog.RunLog.InfofWithCtx(ctx, "%s %s: %s <%3d> (%dms) |%15s |%s |%d", r.Proto, r.Method,
			path, statusRes.Status, latency, clientIP, clientUserAgent, syscall.Getuid())
	}
}

func newResponse(w http.ResponseWriter) *StatusResponseWriter {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		hwlog.RunLog.Warn("http.Hijacker not implement")
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		hwlog.RunLog.Warn("http.Flusher not implement")
	}
	closeNotifier, ok := w.(http.CloseNotifier)
	if !ok {
		hwlog.RunLog.Warn("http.CloseNotifier not implement")
	}
	statusRes := &StatusResponseWriter{
		ResponseWriter: w,
		Hijacker:       hijacker,
		Flusher:        flusher,
		CloseNotifier:  closeNotifier,
		Status:         http.StatusOK,
	}
	return statusRes
}

func initContext(req *http.Request) context.Context {
	ctx := context.Background()
	reqID := req.Header.Get(hwlog.ReqID.String())
	if reqID != "" {
		ctx = context.WithValue(context.Background(), hwlog.ReqID, reqID)
	}
	id := req.Header.Get(hwlog.UserID.String())
	if id != "" {
		ctx = context.WithValue(ctx, hwlog.UserID, id)
	}
	return ctx
}

func returnToken(ctx context.Context, concurrency chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			hwlog.RunLog.Errorf("go routine failed with %#v", err)
		}
	}()
	if concurrency == nil {
		hwlog.RunLog.Error("return token error")
		return
	}
	timer := time.NewTimer(time.Second * second5)
	defer timer.Stop()
	select {
	case _, ok := <-timer.C:
		if !ok {
			return
		}
		concurrency <- struct{}{}
		hwlog.RunLog.Debugf("recover token num：%d", len(concurrency))
		return
	case _, ok := <-ctx.Done():
		err := ctx.Err()
		if !ok || err != nil {
			hwlog.RunLog.Debugf("%+v:%+v", err, ok)
		}
		return
	}
}

// NewLimitHandler new a bucket-token limiter
func NewLimitHandler(maxConcur, maxConcurrency int, handler http.Handler, printLog bool) (http.Handler, error) {
	return NewLimitHandlerWithMethod(maxConcur, maxConcurrency, handler, printLog, "")
}

// NewLimitHandlerWithMethod  new a bucket-token limiter with specific http method
func NewLimitHandlerWithMethod(maxConcur, maxConcurrency int, handler http.Handler, printLog bool,
	httpMethod string) (http.Handler, error) {
	if maxConcur < 1 || maxConcur > maxConcurrency {
		return nil, errors.New("maxConcurrency parameter error")
	}
	conchan := make(chan struct{}, maxConcur)
	cfg := createHandlerConfig{
		ch:            conchan,
		handler:       handler,
		printLog:      printLog,
		httpMethod:    httpMethod,
		bodySizeLimit: DefaultDataLimit,
		expiredTime:   -1,
	}
	return createHandler(cfg), nil
}

type createHandlerConfig struct {
	ch            chan struct{}
	handler       http.Handler
	printLog      bool
	httpMethod    string
	bodySizeLimit int64
	expiredTime   time.Duration
	overdueTime   time.Duration
	cacheSize     int
}

func createHandler(cfg createHandlerConfig) *limitHandler {
	limiter := ConcurrentLimiter{
		concurrency: cfg.ch,
		overdueTime: cfg.overdueTime,
	}
	limiter.Init()
	h := &limitHandler{
		ipLimiterFactory: &TimeWindowLimiterFactory{
			expiredTime: cfg.expiredTime,
		},
		ipLimiters: cache.New(cfg.cacheSize),
		conCurrentHandler: &concHandlerWithCtxAwareLimiter{
			handler: cfg.handler,
			log:     cfg.printLog,
			limiter: &limiter,
		},
		log:        cfg.printLog,
		limitBytes: cfg.bodySizeLimit,
	}
	return h
}

// NewLimitHandlerV2 new a bucket-token limiter which contains limit request by IP
func NewLimitHandlerV2(handler http.Handler, conf *HandlerConfig) (http.Handler, error) {
	if conf == nil {
		return nil, errors.New("parameter error")
	}
	if err := conf.basicConfigCheck(); err != nil {
		return nil, err
	}
	arr0, arr1, err := conf.ipConCurrencyParse()
	if err != nil {
		return nil, err
	}
	cfg := createHandlerConfig{
		ch:            make(chan struct{}, conf.TotalConCurrency),
		handler:       handler,
		printLog:      conf.PrintLog,
		httpMethod:    conf.Method,
		bodySizeLimit: conf.LimitBytes,
		expiredTime:   time.Duration(arr1 * int64(time.Second) / arr0),
		cacheSize:     DefaultCacheSize,
	}
	h := createHandler(cfg)
	return h, nil
}

// NewLimitHandlerV3 add limiters for http.Handler, which includes a shared limiter for all http request
// and message-limiter for each IP address. All limiters are implemented through bucket-token.
func NewLimitHandlerV3(handler http.Handler, conf *HandlerConfigV3) (http.Handler, error) {
	if conf == nil {
		return nil, errors.New("parameter error")
	}
	if err := conf.basicConfigCheck(); err != nil {
		return nil, err
	}
	if conf.IPBurst <= 0 {
		return nil, errors.New("burst config for ip is not allowed")
	}

	// arr0 and arr1 are passed the reg checker, so can not be zero.
	arr0, arr1, err := conf.ipConCurrencyParse()
	if err != nil {
		return nil, err
	}
	cfg := createHandlerConfig{
		ch:            make(chan struct{}, conf.TotalConCurrency),
		handler:       handler,
		printLog:      conf.PrintLog,
		httpMethod:    conf.Method,
		bodySizeLimit: conf.LimitBytes,
		expiredTime:   time.Duration(arr1 * int64(time.Second) / arr0),
		cacheSize:     DefaultCacheSize,
	}
	h := createHandler(cfg)
	h.ipLimiterFactory = &BurstMgsLimiterFactory{
		limit: float64(arr0) / float64(arr1),
		burst: conf.IPBurst,
	}
	return h, nil
}

func (conf *HandlerConfig) basicConfigCheck() error {
	if conf.TotalConCurrency < 1 || conf.TotalConCurrency > defaultMaxConcurrency {
		return errors.New("totalConCurrency parameter error")
	}
	if len(conf.Method) > maxStringLen {
		return errors.New("method parameter error")
	}
	if conf.CacheSize <= 0 {
		hwlog.RunLog.Info("use default cache size")
		conf.CacheSize = DefaultCacheSize
	}
	reg := regexp.MustCompile(IPReqLimitReg)
	if !reg.Match([]byte(conf.IPConCurrency)) {
		return errors.New("IPConCurrency parameter error")
	}
	return nil
}

func (conf *HandlerConfig) ipConCurrencyParse() (int64, int64, error) {
	if conf == nil {
		return 0, 0, errors.New("parameter error")
	}
	arr := strings.Split(conf.IPConCurrency, "/")
	if len(arr) != arrLen || arr[0] == "0" {
		return 0, 0, errors.New("IPConCurrency parameter error")
	}
	arr1, err := strconv.ParseInt(arr[1], 0, 0)
	if err != nil || arr1 == 0 {
		return 0, 0, errors.New("IPConCurrency parameter error, parse to int failed")
	}
	arr0, err := strconv.ParseInt(arr[0], 0, 0)
	if err != nil || arr0 == 0 {
		return 0, 0, errors.New("IPConCurrency parameter error,parse to int failed")
	}
	return arr0, arr1, nil
}

func (h *limitHandler) limitSingleIp(clientIP string) bool {
	value, err := h.ipLimiters.Get(clientIP)
	if err == nil {
		limiter, ok := value.(IndependentLimiter)
		if !ok {
			if h.log {
				hwlog.RunLog.Error("limiter type is invalid")
			}
			return false
		}
		return limiter.Allow()
	}

	if !strings.Contains(err.Error(), "no value found") {
		if h.log {
			hwlog.RunLog.Errorf("ip limiter cache get data failed: %v", err)
		}
		return false
	}

	limiter := h.ipLimiterFactory.Create()
	if setErr := h.ipLimiters.Set(clientIP, limiter, -1); setErr == nil {
		return limiter.Allow()
	}
	if h.log {
		hwlog.RunLog.Errorf("ip liter cache set data failed: %v", err)
	}

	return false
}
