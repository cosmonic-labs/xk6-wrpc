package k6wrpc

import (
	"context"
	"time"
	"xk6-wrpc/internal/xk6/wrpc/blaster"

	uuid "github.com/nu7hatch/gouuid"

	"github.com/grafana/sobek"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
	wrpc "wrpc.io/go"
)

var DefaultBlasterTimeout = 10 * 1000

type wasiBlaster struct {
	vu      modules.VU
	obj     *sobek.Object
	metrics *wrpcMetrics
	tags    *metrics.TagSet
	invoker wrpc.Invoker
}

func newBlaster(vu modules.VU, wm *wrpcMetrics, options clientOptions) (*wasiBlaster, error) {
	rt := vu.Runtime()

	driver, err := newNatsDriver(vu, wm, options.NATS, options.Tags)
	if err != nil {
		return nil, err
	}

	w := &wasiBlaster{
		vu:      vu,
		metrics: wm,
		tags:    wm.extendTagSet(options.Tags),
		obj:     rt.NewObject(),
		invoker: driver.nc,
	}

	if err := w.obj.Set("blast", w.doBlast); err != nil {
		return nil, err
	}

	return w, nil
}

type blasterOptions struct {
	CpuBurnMs    int
	MemoryBurnMb int
	WaitMs       int
	Payload      string
	TimeoutMs    int
}

func (w *wasiBlaster) doBlast(options sobek.Value) error {
	timeout := DefaultBlasterTimeout
	id, _ := uuid.NewV4()
	packet := blaster.Packet{
		Id: id.String(),
	}

	reqStart := time.Now()

	measurements := make([]metrics.Sample, 0)
	defer func() {
		w.metrics.pushIfNotDone(w.vu, measurements...)
	}()

	rt := w.vu.Runtime()
	if options != nil {
		b := blasterOptions{}
		if err := rt.ExportTo(options, &b); err != nil {
			return err
		}
		packet.CpuBurnMs = uint64(b.CpuBurnMs)
		packet.MemBurnMb = uint64(b.MemoryBurnMb)
		packet.WaitMs = uint64(b.WaitMs)
		if b.Payload != "" {
			packet.Payload = []byte(b.Payload)
		}

		if b.TimeoutMs > 0 {
			timeout = b.TimeoutMs
		}
	}

	measurements = append(measurements, w.metrics.sample(w.metrics.blasterOperation, 1, nil))

	ctx, done := context.WithTimeout(w.vu.Context(), time.Duration(timeout)*time.Millisecond)
	defer done()

	err := blaster.Blast(ctx, w.invoker, &packet)
	if err != nil {
		measurements = append(measurements, w.metrics.sample(w.metrics.blasterTransportError, 1, nil))
		return err
	}

	reqDuration := time.Since(reqStart)
	measurements = append(measurements, w.metrics.sample(w.metrics.blasterDuration, metrics.D(reqDuration), nil))

	return nil
}
