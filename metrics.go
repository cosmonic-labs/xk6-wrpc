package k6wrpc

import (
	"time"

	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

type wrpcMetrics struct {
	tags          *metrics.TagSet
	httpOperation *metrics.Metric
	httpError     *metrics.Metric
}

const (
	metricHTTPOperation = "http_operation"
	metricHTTPError     = "http_error"
)

func newWrpcMetrics(registry *metrics.Registry) *wrpcMetrics {
	return &wrpcMetrics{
		httpOperation: registry.MustNewMetric(metricHTTPOperation, metrics.Counter),
		httpError:     registry.MustNewMetric(metricHTTPError, metrics.Counter),
		tags:          registry.RootTagSet(),
	}
}

func (wm *wrpcMetrics) extendTagSet(tags map[string]string) *metrics.TagSet {
	tt := wm.tags

	for k, v := range tags {
		tt = wm.tags.With(k, v)
	}

	return tt
}

func (wm *wrpcMetrics) pushIfNotDone(vu modules.VU, metric *metrics.Metric, value float64, tags *metrics.TagSet) {
	state := vu.State()
	if state == nil {
		return
	}

	ctx := vu.Context()
	if ctx == nil {
		return
	}

	metrics.PushIfNotDone(ctx, state.Samples, metrics.Sample{
		TimeSeries: metrics.TimeSeries{Metric: metric, Tags: tags},
		Time:       time.Now(),
		Value:      value,
	})
}
