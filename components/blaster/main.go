package main

import (
	"blaster-component/internal/xk6/wrpc/blaster"
	"time"
)

//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --world server --out internal wit

func init() {
	blaster.Exports.Blast = blast
}

func blast(pkt blaster.Packet) {
	// Allocate & hold memory during each invocation
	var mem []byte
	if pkt.MemBurnMb > 0 {
		mem = make([]byte, pkt.MemBurnMb*1024*1024)
		_ = mem
	}

	// Simple sleep
	if pkt.WaitMs > 0 {
		<-time.After(time.Duration(pkt.WaitMs) * time.Millisecond)
	}

	// Spin CPU
	if pkt.CPUBurnMs > 0 {
		// NOTE(lxf): this is a busy loop, it will burn CPU
		// Getting creative here cause TinyGo 0.34 seem to have issues with timers in wasm.
		start := time.Now().UnixMilli()
		for now := time.Now().UnixMilli(); now < start+int64(pkt.CPUBurnMs); now = time.Now().UnixMilli() {
			_ = time.Now()
		}
	}
}

func main() {}
