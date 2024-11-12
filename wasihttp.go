package k6wrpc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
	wasitypes "xk6-wrpc/internal/wasi/http/types"
	"xk6-wrpc/internal/wrpc/http/incoming_handler"
	wrpctypes "xk6-wrpc/internal/wrpc/http/types"

	"github.com/grafana/sobek"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/lib/netext/httpext"
	"go.k6.io/k6/metrics"
	wrpc "wrpc.io/go"
)

var ErrRPC = errors.New("rpc error")

// default timeout in ms
var DefaultHTTPTimeout = int64(30 * 1000)

type wasiHTTP struct {
	vu               modules.VU
	obj              *sobek.Object
	metrics          *wrpcMetrics
	tags             *metrics.TagSet
	invoker          wrpc.Invoker
	responseCallback func(int) bool
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
		invoker: driver.nc,
		responseCallback: func(status int) bool {
			return status <= 200 && status < 300
		},
	}

	if err := w.obj.Set("get", w.noBodyRequest(http.MethodGet)); err != nil {
		return nil, err
	}
	if err := w.obj.Set("head", w.noBodyRequest(http.MethodGet)); err != nil {
		return nil, err
	}

	if err := w.obj.Set("del", w.bodyRequest(http.MethodDelete)); err != nil {
		return nil, err
	}
	if err := w.obj.Set("options", w.noBodyRequest(http.MethodOptions)); err != nil {
		return nil, err
	}
	if err := w.obj.Set("patch", w.noBodyRequest(http.MethodPatch)); err != nil {
		return nil, err
	}
	if err := w.obj.Set("post", w.bodyRequest(http.MethodPost)); err != nil {
		return nil, err
	}
	if err := w.obj.Set("put", w.bodyRequest(http.MethodPut)); err != nil {
		return nil, err
	}

	return w, nil
}

type httpResponse struct {
	Status  int
	Headers map[string][]string
	Body    []byte
}

func (w *wasiHTTP) noBodyRequest(method string) func(url sobek.Value, args ...sobek.Value) (*httpResponse, error) {
	return func(url sobek.Value, args ...sobek.Value) (*httpResponse, error) {
		args = append([]sobek.Value{sobek.Undefined()}, args...)
		return w.request(method, url, args...)
	}
}

func (w *wasiHTTP) bodyRequest(method string) func(url sobek.Value, args ...sobek.Value) (*httpResponse, error) {
	return func(url sobek.Value, args ...sobek.Value) (*httpResponse, error) {
		return w.request(method, url, args...)
	}
}

type wasiTrailer struct{}

func (w wasiTrailer) Receive() ([]*wrpc.Tuple2[string, [][]byte], error) {
	ret := make([]*wrpc.Tuple2[string, [][]byte], 0)
	return ret, nil
}

func (w wasiTrailer) Close() error {
	return nil
}

func jsBodyToWrpc(body interface{}) (io.ReadCloser, error) {
	switch data := body.(type) {
	case string:
		return io.NopCloser(bytes.NewBufferString(data)), nil
	case []byte:
		return io.NopCloser(bytes.NewBuffer(data)), nil
	case sobek.ArrayBuffer:
		return io.NopCloser(bytes.NewBuffer(data.Bytes())), nil
	case map[string]interface{}:
		d, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewBuffer(d)), nil
	case nil:
		return http.NoBody, nil
	default:
		return nil, fmt.Errorf("unsupported body type %T", body)
	}
}

func (w *wasiHTTP) request(method string, url sobek.Value, args ...sobek.Value) (*httpResponse, error) {
	timeout := DefaultHTTPTimeout
	consumeBody := false
	reqStart := time.Now()

	measurements := make([]metrics.Sample, 0)
	defer func() {
		w.metrics.pushIfNotDone(w.vu, measurements...)
	}()

	parsedURL, err := httpext.ToURL(url.Export())
	if err != nil {
		return nil, err
	}
	u := parsedURL.GetURL()

	headers := make([]*wrpc.Tuple2[string, [][]uint8], 0)

	var body io.ReadCloser

	var trailers wrpc.Receiver[[]*wrpc.Tuple2[string, [][]uint8]]
	trailers = wasiTrailer{}

	bodyParam, params := splitRequestArgs(args)
	body, err = jsBodyToWrpc(bodyParam.Export())
	if err != nil {
		return nil, err
	}

	if params != nil {
		p := params.Export().(map[string]interface{})

		// auth
		if data, ok := p["auth"]; ok {
			d := data.(map[string]interface{})

			if user, ok := d["username"]; ok {
				if pass, ok := d["password"]; ok {
					headers = append(headers, &wrpc.Tuple2[string, [][]uint8]{
						V0: "Authorization",
						V1: [][]uint8{
							[]byte(
								fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pass))))),
						},
					})
				}
			}

			// TODO(lxf): bearer tokens
		}

		// timeout in ms
		if data, ok := p["timeout"]; ok {
			timeout = data.(int64)
		}

		// if we should read the body or not
		if data, ok := p["consume"]; ok {
			consumeBody = data.(bool)
		}

		// headers
		if data, ok := p["headers"]; ok {
			h := data.(map[string]interface{})
			for k, v := range h {
				vs := v.(string)
				headers = append(headers, &wrpc.Tuple2[string, [][]uint8]{
					V0: k,
					V1: [][]uint8{[]byte(vs)},
				})
			}
		}

	}

	pathWithQuery := u.RequestURI()
	authority := u.Host

	wreq := &wrpctypes.Request{
		Headers:       headers,
		Method:        HttpMethodToWrpc(method),
		Scheme:        HttpSchemeToWrpc(u.Scheme),
		PathWithQuery: &pathWithQuery,
		Authority:     &authority,
		Body:          body,
		// TODO(lxf): implement trailers?
		Trailers: trailers,
	}

	ctx, done := context.WithTimeout(w.vu.Context(), time.Duration(timeout)*time.Millisecond)
	defer done()

	measurements = append(measurements, w.metrics.sample(w.metrics.httpRequest, 1, nil))

	res, _, err := incoming_handler.Handle(ctx, w.invoker, wreq)
	if err != nil {
		measurements = append(measurements, w.metrics.sample(w.metrics.transportError, 1, nil))
		return nil, err
	}

	if res.Err != nil {
		measurements = append(measurements, w.metrics.sample(w.metrics.httpError, 1, nil))
		return nil, res.Err
	}

	resp := res.Ok

	incomingBody := bytes.NewBuffer(nil)
	if consumeBody {
		if _, err := io.Copy(incomingBody, resp.Body); err != nil {
			return nil, err
		}
		resp.Body.Close()
	}

	reqDuration := time.Since(reqStart)
	measurements = append(measurements, w.metrics.sample(w.metrics.httpDuration, metrics.D(reqDuration), nil))

	var responseCallback func(int) bool
	responseCallback = w.responseCallback

	if responseCallback(int(resp.Status)) {
		measurements = append(measurements, w.metrics.sample(w.metrics.httpResponse, 1, nil))
	} else {
		measurements = append(measurements, w.metrics.sample(w.metrics.httpInvalidResponse, 1, nil))
	}

	incomingHeaders := make(http.Header)
	for _, header := range resp.Headers {
		for _, v := range header.V1 {
			incomingHeaders.Add(header.V0, string(v))
		}
	}

	return &httpResponse{
		Status:  int(resp.Status),
		Headers: incomingHeaders,
		Body:    incomingBody.Bytes(),
	}, nil
}

func xinit() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})))
}

func splitRequestArgs(args []sobek.Value) (body sobek.Value, params sobek.Value) {
	if len(args) > 0 {
		body = args[0]
	}
	if len(args) > 1 {
		params = args[1]
	}
	return body, params
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
