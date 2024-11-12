package k6wrpc

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/grafana/sobek"
	"github.com/nats-io/nats.go"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/lib/netext"
)

// RootModule is the global module object type. It is instantiated once per test
// run and will be used to create WRPC module instances for each VU.
//
// TODO: add sync.Once for all of the deprecation warnings we might want to do
// for the old k6/http APIs here, so they are shown only once in a test run.
type RootModule struct{}

// ModuleInstance represents an instance of the WRPC module for every VU.
type ModuleInstance struct {
	vu         modules.VU
	rootModule *RootModule
	exports    *sobek.Object
	metrics    *wrpcMetrics
}

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &ModuleInstance{}
)

// New returns a pointer to a new WRPC RootModule.
func New() *RootModule {
	return &RootModule{}
}

// NewModuleInstance returns an WRPC module instance for each VU.
func (r *RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	rt := vu.Runtime()

	env := vu.InitEnv()
	if env == nil {
		common.Throw(rt, fmt.Errorf("no environment found"))
		return nil
	}

	registry := env.Registry
	if registry == nil {
		common.Throw(rt, fmt.Errorf("missing registry"))
		return nil
	}

	mi := &ModuleInstance{
		vu:         vu,
		rootModule: r,
		metrics:    newWrpcMetrics(registry),
		exports:    rt.NewObject(),
	}
	mi.defineConstants()

	mustExport := func(name string, value interface{}) {
		if err := mi.exports.Set(name, value); err != nil {
			common.Throw(rt, err)
		}
	}

	mustExport("http", mi.httpClient)

	return mi
}

type clientOptions struct {
	Tags map[string]string `json:"tags,omitempty"`
	NATS *natsClientOption `json:"nats,omitempty"`
}

func (mi *ModuleInstance) httpClient(rawOptions *sobek.Object) *sobek.Object {
	rt := mi.vu.Runtime()

	data, err := rawOptions.MarshalJSON()
	if err != nil {
		common.Throw(rt, err)
		return nil
	}

	options := clientOptions{
		Tags: make(map[string]string),
	}
	if err := json.Unmarshal(data, &options); err != nil {
		common.Throw(rt, err)
		return nil
	}

	w, err := newWasiHTTP(mi.vu, mi.metrics, options)
	if err != nil {
		common.Throw(rt, err)
		return nil
	}

	return w.obj
}

// Exports returns the JS values this module exports.
func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Default: mi.exports,
	}
}

func (mi *ModuleInstance) defineConstants() {
	rt := mi.vu.Runtime()
	mustAddProp := func(name, val string) {
		err := mi.exports.DefineDataProperty(
			name, rt.ToValue(val), sobek.FLAG_FALSE, sobek.FLAG_FALSE, sobek.FLAG_TRUE,
		)
		if err != nil {
			common.Throw(rt, err)
		}
	}
	mustAddProp("TLS_1_0", netext.TLS_1_0)
	mustAddProp("TLS_1_1", netext.TLS_1_1)
	mustAddProp("TLS_1_2", netext.TLS_1_2)
	mustAddProp("TLS_1_3", netext.TLS_1_3)
	mustAddProp("OCSP_STATUS_GOOD", netext.OCSP_STATUS_GOOD)
	mustAddProp("OCSP_STATUS_REVOKED", netext.OCSP_STATUS_REVOKED)
	mustAddProp("OCSP_STATUS_SERVER_FAILED", netext.OCSP_STATUS_SERVER_FAILED)
	mustAddProp("OCSP_STATUS_UNKNOWN", netext.OCSP_STATUS_UNKNOWN)
	mustAddProp("OCSP_REASON_UNSPECIFIED", netext.OCSP_REASON_UNSPECIFIED)
	mustAddProp("OCSP_REASON_KEY_COMPROMISE", netext.OCSP_REASON_KEY_COMPROMISE)
	mustAddProp("OCSP_REASON_CA_COMPROMISE", netext.OCSP_REASON_CA_COMPROMISE)
	mustAddProp("OCSP_REASON_AFFILIATION_CHANGED", netext.OCSP_REASON_AFFILIATION_CHANGED)
	mustAddProp("OCSP_REASON_SUPERSEDED", netext.OCSP_REASON_SUPERSEDED)
	mustAddProp("OCSP_REASON_CESSATION_OF_OPERATION", netext.OCSP_REASON_CESSATION_OF_OPERATION)
	mustAddProp("OCSP_REASON_CERTIFICATE_HOLD", netext.OCSP_REASON_CERTIFICATE_HOLD)
	mustAddProp("OCSP_REASON_REMOVE_FROM_CRL", netext.OCSP_REASON_REMOVE_FROM_CRL)
	mustAddProp("OCSP_REASON_PRIVILEGE_WITHDRAWN", netext.OCSP_REASON_PRIVILEGE_WITHDRAWN)
	mustAddProp("OCSP_REASON_AA_COMPROMISE", netext.OCSP_REASON_AA_COMPROMISE)
}

type Client struct {
	moduleInstance   *ModuleInstance
	responseCallback func(int) bool

	nc         *nats.Conn
	ncOnce     sync.Once
	natsURL    string
	natsPrefix string
}

func init() {
	modules.Register("k6/x/wrpc", new(RootModule))
}
