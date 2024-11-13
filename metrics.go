package k6wrpc

import (
	"time"

	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

type wrpcMetrics struct {
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

	// operations
	blasterOperation *metrics.Metric
	// underlying wrpc encoding errors
	blasterTransportError *metrics.Metric
	// operation duration
	blasterDuration *metrics.Metric
}

const (
	metricHTTPRequest         = "wrpc_http_request"
	metricHTTPResponse        = "wrpc_http_response"
	metricHTTPInvalidResponse = "wrpc_http_invalid_response"
	metricHTTPError           = "wrpc_http_error"
	metricTransportError      = "wrpc_transport_error"
	metricHTTPDuration        = "wrpc_http_duration"

	metriBlasterOperation       = "wrpc_blaster_operation"
	metricBlasterTransportError = "wrpc_blaster_transport_error"
	metricBlasterDuration       = "wrpc_blaster_duration"
)

func newWrpcMetrics(registry *metrics.Registry) *wrpcMetrics {
	return &wrpcMetrics{
		httpRequest:         registry.MustNewMetric(metricHTTPRequest, metrics.Counter),
		httpResponse:        registry.MustNewMetric(metricHTTPResponse, metrics.Counter),
		httpDuration:        registry.MustNewMetric(metricHTTPDuration, metrics.Trend, metrics.Time),
		httpInvalidResponse: registry.MustNewMetric(metricHTTPInvalidResponse, metrics.Counter),
		httpError:           registry.MustNewMetric(metricHTTPError, metrics.Counter),
		transportError:      registry.MustNewMetric(metricTransportError, metrics.Counter),

		blasterOperation:      registry.MustNewMetric(metriBlasterOperation, metrics.Counter),
		blasterTransportError: registry.MustNewMetric(metricBlasterTransportError, metrics.Counter),
		blasterDuration:       registry.MustNewMetric(metricBlasterDuration, metrics.Trend, metrics.Time),
	}
}

func (wm *wrpcMetrics) sample(metric *metrics.Metric, value float64, tags *metrics.TagSet) metrics.Sample {
	return metrics.Sample{
		TimeSeries: metrics.TimeSeries{Metric: metric, Tags: tags},
		Time:       time.Now(),
		Value:      value,
	}
}

func (wm *wrpcMetrics) pushIfNotDone(vu modules.VU, samples ...metrics.Sample) {
	metrics.PushIfNotDone(vu.Context(), vu.State().Samples, metrics.Samples(samples))
}
