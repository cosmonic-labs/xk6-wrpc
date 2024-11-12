package k6wrpc

import (
	"github.com/nats-io/nats.go"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
	wrpcnats "wrpc.io/go/nats"
)

type natsClientOption struct {
	URL    string `json:"url"`
	Prefix string `json:"prefix,omitempty"`
}

type natsDriver struct {
	nc   *wrpcnats.Client
	tags *metrics.TagSet
}

func newNatsDriver(vu modules.VU, wm *wrpcMetrics, options *natsClientOption, tags map[string]string) (*natsDriver, error) {
	nc, err := nats.Connect(options.URL)
	if err != nil {
		return nil, err
	}
	client := &natsDriver{
		nc:   wrpcnats.NewClient(nc, wrpcnats.WithPrefix(options.Prefix)),
		tags: wm.extendTagSet(tags),
	}
	return client, nil
}
