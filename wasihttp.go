package k6wrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	wasitypes "xk6-wrpc/internal/wasi/http/types"
	"xk6-wrpc/internal/wrpc/http/incoming_handler"
	wrpctypes "xk6-wrpc/internal/wrpc/http/types"

	"github.com/grafana/sobek"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
	wrpc "wrpc.io/go"
)

var ErrRPC = errors.New("rpc error")

type wasiHTTP struct {
	vu      modules.VU
	obj     *sobek.Object
	metrics *wrpcMetrics
	tags    *metrics.TagSet
	client  *http.Client
}

func newWasiHTTP(vu modules.VU, wm *wrpcMetrics, options clientOptions) (*wasiHTTP, error) {
	rt := vu.Runtime()

	driver, err := newNatsDriver(vu, wm, options.NATS, options.Tags)
	if err != nil {
		return nil, err
	}

	w := &wasiHTTP{
		vu:      vu,
		metrics: wm,
		tags:    wm.extendTagSet(options.Tags),
		obj:     rt.NewObject(),
		client: &http.Client{
			Transport: &wasiRoundTripper{
				driver:  driver.nc,
				invoker: incoming_handler.Handle,
			},
		},
	}

	if err := w.obj.Set("get", rt.ToValue(w.httpGet)); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *wasiHTTP) httpGet(url string) {
	resp, err := w.client.Get(url)

	w.metrics.pushIfNotDone(w.vu, w.metrics.httpOperation, 1, w.tags)
	if err != nil {
		w.metrics.pushIfNotDone(w.vu, w.metrics.httpError, 1, w.tags)
		return
	}

	if resp.StatusCode > 399 && resp.StatusCode < 600 {
		w.metrics.pushIfNotDone(w.vu, w.metrics.httpError, 1, w.tags)
	}
}

var _ http.RoundTripper = (*wasiRoundTripper)(nil)

type IncomingHandlerOption func(*wasiRoundTripper)

type wasiRoundTripper struct {
	driver wrpc.Invoker
	// NOTE(lxf): to override during tests
	invoker func(context.Context, wrpc.Invoker, *wrpctypes.Request) (*wrpc.Result[incoming_handler.Response, incoming_handler.ErrorCode], <-chan error, error)
}

func (p *wasiRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	outgoingBodyTrailer := HttpBodyToWrpc(r.Body, r.Trailer)
	pathWithQuery := r.URL.Path
	if r.URL.RawQuery != "" {
		pathWithQuery += "?" + r.URL.RawQuery
	}
	wreq := &wrpctypes.Request{
		Headers:       HttpHeaderToWrpc(r.Header),
		Method:        HttpMethodToWrpc(r.Method),
		Scheme:        HttpSchemeToWrpc(r.URL.Scheme),
		PathWithQuery: &pathWithQuery,
		Authority:     &r.Host,
		Body:          outgoingBodyTrailer,
		Trailers:      outgoingBodyTrailer,
	}

	wresp, errCh, err := p.invoker(r.Context(), p.driver, wreq)
	if err != nil {
		return nil, err
	}

	if wresp.Err != nil {
		return nil, fmt.Errorf("%w: %s", ErrRPC, wresp.Err)
	}

	respBody, trailers := WrpcBodyToHttp(wresp.Ok.Body, wresp.Ok.Trailers)

	resp := &http.Response{
		StatusCode: int(wresp.Ok.Status),
		Header:     make(http.Header),
		Request:    r,
		Body:       respBody,
		Trailer:    trailers,
	}

	for _, hdr := range wresp.Ok.Headers {
		for _, hdrVal := range hdr.V1 {
			resp.Header.Add(hdr.V0, string(hdrVal))
		}
	}

	errList := []error{}
	for err := range errCh {
		errList = append(errList, err)
	}

	if len(errList) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrRPC, errList)
	}

	return resp, nil
}

type wrpcIncomingBody struct {
	body           io.Reader
	trailer        http.Header
	trailerRx      wrpc.Receiver[[]*wrpc.Tuple2[string, [][]byte]]
	trailerOnce    sync.Once
	trailerIsReady uint32
}

func (r *wrpcIncomingBody) Close() error {
	return nil
}

func (r *wrpcIncomingBody) readTrailerOnce() {
	r.trailerOnce.Do(func() {
		trailers, err := r.trailerRx.Receive()
		if err != nil {
			return
		}
		for _, header := range trailers {
			for _, value := range header.V1 {
				r.trailer.Add(header.V0, string(value))
			}
		}
		atomic.CompareAndSwapUint32(&r.trailerIsReady, 0, 1)
	})
}

func (r *wrpcIncomingBody) Read(b []byte) (int, error) {
	n, err := r.body.Read(b)
	if err == io.EOF {
		r.readTrailerOnce()
	}
	return n, err
}

type wrpcOutgoingBody struct {
	body        io.ReadCloser
	trailer     http.Header
	bodyIsDone  chan struct{}
	trailerOnce sync.Once
}

func (r *wrpcOutgoingBody) Read(b []byte) (int, error) {
	n, err := r.body.Read(b)
	if err == io.EOF {
		r.finish()
	}
	return n, err
}

func (r *wrpcOutgoingBody) Receive() ([]*wrpc.Tuple2[string, [][]byte], error) {
	<-r.bodyIsDone
	trailers := HttpHeaderToWrpc(r.trailer)
	return trailers, nil
}

func (r *wrpcOutgoingBody) finish() {
	r.trailerOnce.Do(func() {
		r.body.Close()
		close(r.bodyIsDone)
	})
}

func (r *wrpcOutgoingBody) Close() error {
	r.finish()

	return nil
}

func HttpBodyToWrpc(body io.ReadCloser, trailer http.Header) *wrpcOutgoingBody {
	if body == nil {
		body = http.NoBody
	}
	return &wrpcOutgoingBody{
		body:       body,
		trailer:    trailer,
		bodyIsDone: make(chan struct{}, 1),
	}
}

func WrpcBodyToHttp(body io.Reader, trailerRx wrpc.Receiver[[]*wrpc.Tuple2[string, [][]uint8]]) (*wrpcIncomingBody, http.Header) {
	trailer := make(http.Header)
	return &wrpcIncomingBody{
		body:      body,
		trailerRx: trailerRx,
		trailer:   trailer,
	}, trailer
}

func HttpMethodToWrpc(method string) *wrpctypes.Method {
	switch method {
	case http.MethodConnect:
		return wasitypes.NewMethodConnect()
	case http.MethodGet:
		return wasitypes.NewMethodGet()
	case http.MethodHead:
		return wasitypes.NewMethodHead()
	case http.MethodPost:
		return wasitypes.NewMethodPost()
	case http.MethodPut:
		return wasitypes.NewMethodPut()
	case http.MethodPatch:
		return wasitypes.NewMethodPatch()
	case http.MethodDelete:
		return wasitypes.NewMethodDelete()
	case http.MethodOptions:
		return wasitypes.NewMethodOptions()
	case http.MethodTrace:
		return wasitypes.NewMethodTrace()
	default:
		return wasitypes.NewMethodOther(method)
	}
}

func HttpSchemeToWrpc(scheme string) *wrpctypes.Scheme {
	switch scheme {
	case "http":
		return wasitypes.NewSchemeHttp()
	case "https":
		return wasitypes.NewSchemeHttps()
	default:
		return wasitypes.NewSchemeOther(scheme)
	}
}

func HttpHeaderToWrpc(header http.Header) []*wrpc.Tuple2[string, [][]uint8] {
	wasiHeader := make([]*wrpc.Tuple2[string, [][]uint8], 0, len(header))
	for k, vals := range header {
		var uintVals [][]uint8
		for _, v := range vals {
			uintVals = append(uintVals, []byte(v))
		}
		wasiHeader = append(wasiHeader, &wrpc.Tuple2[string, [][]uint8]{
			V0: k,
			V1: uintVals,
		})
	}

	return wasiHeader
}
