// Generated by `wit-bindgen-wrpc-go` 0.11.0. DO NOT EDIT!
package blaster

import (
	bytes "bytes"
	context "context"
	binary "encoding/binary"
	errors "errors"
	fmt "fmt"
	io "io"
	slog "log/slog"
	math "math"
	sync "sync"
	atomic "sync/atomic"
	wrpc "wrpc.io/go"
)

type Packet struct {
	// The ID of the packet
	Id string
	// The payload of the packet
	Payload []uint8
	// Tells the component to allocate memory
	MemBurnMb uint64
	// Tells the component to spinlock the CPU
	CpuBurnMs uint64
	// Tells the component to sleep
	WaitMs uint64
}

func (v *Packet) String() string { return "Packet" }

func (v *Packet) WriteToIndex(w wrpc.ByteWriter) (func(wrpc.IndexWriter) error, error) {
	writes := make(map[uint32]func(wrpc.IndexWriter) error, 5)
	slog.Debug("writing field", "name", "id")
	write0, err := (func(wrpc.IndexWriter) error)(nil), func(v string, w io.Writer) (err error) {
		n := len(v)
		if n > math.MaxUint32 {
			return fmt.Errorf("string byte length of %d overflows a 32-bit integer", n)
		}
		if err = func(v int, w io.Writer) error {
			b := make([]byte, binary.MaxVarintLen32)
			i := binary.PutUvarint(b, uint64(v))
			slog.Debug("writing string byte length", "len", n)
			_, err = w.Write(b[:i])
			return err
		}(n, w); err != nil {
			return fmt.Errorf("failed to write string byte length of %d: %w", n, err)
		}
		slog.Debug("writing string bytes")
		_, err = w.Write([]byte(v))
		if err != nil {
			return fmt.Errorf("failed to write string bytes: %w", err)
		}
		return nil
	}(v.Id, w)
	if err != nil {
		return nil, fmt.Errorf("failed to write `id` field: %w", err)
	}
	if write0 != nil {
		writes[0] = write0
	}
	slog.Debug("writing field", "name", "payload")
	write1, err := func(v []uint8, w interface {
		io.ByteWriter
		io.Writer
	}) (write func(wrpc.IndexWriter) error, err error) {
		n := len(v)
		if n > math.MaxUint32 {
			return nil, fmt.Errorf("list length of %d overflows a 32-bit integer", n)
		}
		if err = func(v int, w io.Writer) error {
			b := make([]byte, binary.MaxVarintLen32)
			i := binary.PutUvarint(b, uint64(v))
			slog.Debug("writing list length", "len", n)
			_, err = w.Write(b[:i])
			return err
		}(n, w); err != nil {
			return nil, fmt.Errorf("failed to write list length of %d: %w", n, err)
		}
		slog.Debug("writing list elements")
		writes := make(map[uint32]func(wrpc.IndexWriter) error, n)
		for i, e := range v {
			write, err := (func(wrpc.IndexWriter) error)(nil), func(v uint8, w io.ByteWriter) error {
				slog.Debug("writing u8 byte")
				return w.WriteByte(v)
			}(e, w)
			if err != nil {
				return nil, fmt.Errorf("failed to write list element %d: %w", i, err)
			}
			if write != nil {
				writes[uint32(i)] = write
			}
		}
		if len(writes) > 0 {
			return func(w wrpc.IndexWriter) error {
				var wg sync.WaitGroup
				var wgErr atomic.Value
				for index, write := range writes {
					wg.Add(1)
					w, err := w.Index(index)
					if err != nil {
						return fmt.Errorf("failed to index nested list writer: %w", err)
					}
					write := write
					go func() {
						defer wg.Done()
						if err := write(w); err != nil {
							wgErr.Store(err)
						}
					}()
				}
				wg.Wait()
				err := wgErr.Load()
				if err == nil {
					return nil
				}
				return err.(error)
			}, nil
		}
		return nil, nil
	}(v.Payload, w)
	if err != nil {
		return nil, fmt.Errorf("failed to write `payload` field: %w", err)
	}
	if write1 != nil {
		writes[1] = write1
	}
	slog.Debug("writing field", "name", "mem-burn-mb")
	write2, err := (func(wrpc.IndexWriter) error)(nil), func(v uint64, w io.Writer) (err error) {
		b := make([]byte, binary.MaxVarintLen64)
		i := binary.PutUvarint(b, uint64(v))
		slog.Debug("writing u64")
		_, err = w.Write(b[:i])
		return err
	}(v.MemBurnMb, w)
	if err != nil {
		return nil, fmt.Errorf("failed to write `mem-burn-mb` field: %w", err)
	}
	if write2 != nil {
		writes[2] = write2
	}
	slog.Debug("writing field", "name", "cpu-burn-ms")
	write3, err := (func(wrpc.IndexWriter) error)(nil), func(v uint64, w io.Writer) (err error) {
		b := make([]byte, binary.MaxVarintLen64)
		i := binary.PutUvarint(b, uint64(v))
		slog.Debug("writing u64")
		_, err = w.Write(b[:i])
		return err
	}(v.CpuBurnMs, w)
	if err != nil {
		return nil, fmt.Errorf("failed to write `cpu-burn-ms` field: %w", err)
	}
	if write3 != nil {
		writes[3] = write3
	}
	slog.Debug("writing field", "name", "wait-ms")
	write4, err := (func(wrpc.IndexWriter) error)(nil), func(v uint64, w io.Writer) (err error) {
		b := make([]byte, binary.MaxVarintLen64)
		i := binary.PutUvarint(b, uint64(v))
		slog.Debug("writing u64")
		_, err = w.Write(b[:i])
		return err
	}(v.WaitMs, w)
	if err != nil {
		return nil, fmt.Errorf("failed to write `wait-ms` field: %w", err)
	}
	if write4 != nil {
		writes[4] = write4
	}

	if len(writes) > 0 {
		return func(w wrpc.IndexWriter) error {
			var wg sync.WaitGroup
			var wgErr atomic.Value
			for index, write := range writes {
				wg.Add(1)
				w, err := w.Index(index)
				if err != nil {
					return fmt.Errorf("failed to index nested record writer: %w", err)
				}
				write := write
				go func() {
					defer wg.Done()
					if err := write(w); err != nil {
						wgErr.Store(err)
					}
				}()
			}
			wg.Wait()
			err := wgErr.Load()
			if err == nil {
				return nil
			}
			return err.(error)
		}, nil
	}
	return nil, nil
}
func Blast(ctx__ context.Context, wrpc__ wrpc.Invoker, packet *Packet) (err__ error) {
	var buf__ bytes.Buffer
	write0__, err__ := (packet).WriteToIndex(&buf__)
	if err__ != nil {
		err__ = fmt.Errorf("failed to write `packet` parameter: %w", err__)
		return
	}
	if write0__ != nil {
		err__ = errors.New("unexpected deferred write for synchronous `packet` parameter")
		return
	}
	var w__ wrpc.IndexWriteCloser
	var r__ wrpc.IndexReadCloser
	w__, r__, err__ = wrpc__.Invoke(ctx__, "xk6:wrpc/blaster@0.0.1", "blast", buf__.Bytes())
	if err__ != nil {
		err__ = fmt.Errorf("failed to invoke `blast`: %w", err__)
		return
	}
	defer func() {
		if err := r__.Close(); err != nil {
			slog.ErrorContext(ctx__, "failed to close reader", "instance", "xk6:wrpc/blaster@0.0.1", "name", "blast", "err", err)
		}
	}()
	if cErr__ := w__.Close(); cErr__ != nil {
		slog.DebugContext(ctx__, "failed to close outgoing stream", "instance", "xk6:wrpc/blaster@0.0.1", "name", "blast", "err", cErr__)
	}
	return
}
