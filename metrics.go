package k6wrpc

import (
	"time"

	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

type wrpcMetrics struct {
	tags *metrics.TagSet

	// underlying wrpc encoding errors
	transportError *metrics.Metric
	// requests created
	httpRequest *metrics.Metric
	// valid responses
	httpResponse *metrics.Metric
	// invalid responses
	httpInvalidResponse *metrics.Metric
	// requests that failed to be serviced by the server
	httpError *metrics.Metric
	// http request x response duration
	httpDuration *metrics.Metric
}

const (
	metricHTTPRequest         = "wrpc_http_request"
	metricHTTPResponse        = "wrpc_http_response"
	metricHTTPInvalidResponse = "wrpc_http_invalid_response"
	metricHTTPError           = "wrpc_http_error"
	metricTransportError      = "wrpc_transport_error"

	metricHTTPDuration = "wrpc_http_duration"
)

func newWrpcMetrics(registry *metrics.Registry) *wrpcMetrics {
	return &wrpcMetrics{
		httpRequest:         registry.MustNewMetric(metricHTTPRequest, metrics.Counter),
		httpResponse:        registry.MustNewMetric(metricHTTPResponse, metrics.Counter),
		httpDuration:        registry.MustNewMetric(metricHTTPDuration, metrics.Trend, metrics.Time),
		httpInvalidResponse: registry.MustNewMetric(metricHTTPInvalidResponse, metrics.Counter),
		httpError:           registry.MustNewMetric(metricHTTPError, metrics.Counter),
		transportError:      registry.MustNewMetric(metricTransportError, metrics.Counter),
		tags:                registry.RootTagSet(),
	}
}

func (wm *wrpcMetrics) extendTagSet(tags map[string]string) *metrics.TagSet {
	tt := wm.tags

	if tags != nil {
		for k, v := range tags {
			tt = wm.tags.With(k, v)
		}
	}

	return tt
}

func (wm *wrpcMetrics) sample(metric *metrics.Metric, value float64, tags map[string]string) metrics.Sample {
	return metrics.Sample{
		TimeSeries: metrics.TimeSeries{Metric: metric, Tags: wm.extendTagSet(tags)},
		Time:       time.Now(),
		Value:      value,
	}
}

func (wm *wrpcMetrics) pushIfNotDone(vu modules.VU, samples ...metrics.Sample) {
	state := vu.State()
	if state == nil {
		return
	}

	ctx := vu.Context()
	if ctx == nil {
		return
	}

	metrics.PushIfNotDone(ctx, state.Samples, metrics.Samples(samples))
}
