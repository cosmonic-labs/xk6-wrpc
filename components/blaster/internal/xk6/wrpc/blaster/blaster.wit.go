// Code generated by wit-bindgen-go. DO NOT EDIT.

// Package blaster represents the exported interface "xk6:wrpc/blaster@0.0.1".
package blaster

import (
	"go.bytecodealliance.org/cm"
)

// Packet represents the record "xk6:wrpc/blaster@0.0.1#packet".
//
//	record packet {
//		id: string,
//		payload: list<u8>,
//		mem-burn-mb: u64,
//		cpu-burn-ms: u64,
//		wait-ms: u64,
//	}
type Packet struct {
	_ cm.HostLayout
	// The ID of the packet
	ID string

	// The payload of the packet
	Payload cm.List[uint8]

	// Tells the component to allocate memory
	MemBurnMb uint64

	// Tells the component to spinlock the CPU
	CPUBurnMs uint64

	// Tells the component to sleep
	WaitMs uint64
}
