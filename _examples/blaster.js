import wrpc from "k6/x/wrpc";

export const options = {
  scenarios: {
    contacts: {
      executor: "constant-vus",
      vus: 1,
      duration: "10s",
    },
  },
};

let blaster = wrpc.blaster({
  nats: {
    url: "nats://localhost:4222",
    prefix: "wasmtime",
  },
  tags: { scenario: "contacts" },
});

export default function () {
  // simple roundtrip
  blaster.blast();

  // large packet
  blaster.blast({ payload: "x".repeat(1024 * 1024) });

  // all options
  blaster.blast({
    // tell the component to cpu spin for 100ms
    cpu_burn_ms: 100,
    // tell the component to allocate 10 mb
    memory_burn_mb: 10,
    // tell the component to sleep for 100ms
    wait_ms: 100,
    // include an arbitraty payload in the wrpc message
    payload: "hello world",
    // set the request timeout to 10 seconds
    timeout_ms: 10000,
  });
}
