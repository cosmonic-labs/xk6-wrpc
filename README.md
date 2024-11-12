# wRPC extension for [k6 load testing tool](https://k6.io/)

🚧 **This is a work in progress.** 🚧

What's included:

- `wasi:http` wRPC client
- Generic wRPC blaster client
- Component for receiving blaster packets

To build a new `k6` binary:

```sh
make
```

See examples under [\_examples](./_examples) directory.

## Features

- [x] NATS Transport
  - [ ] Transport Options
  - [ ] Metrics
- [ ] TCP Transport
  - [ ] Transport Options
  - [ ] Metrics
- HTTP Interface
  - [x] Requests with no body
  - [x] Requests with body
  - [x] Ignore Responses
  - [x] Callback for Responses
  - [ ] Asynchronous Requests
  - [x] Metrics
- Load test specific Interface
  - [x] CPU Burn
  - [x] Memory Burn
  - [x] Sleep
  - [x] Payload
  - [x] Metrics

## Blaster API

To be used together with [wrpc-wasmtime-nats](https://github.com/bytecodealliance/wrpc).

Use the [blaster-component](./blaster-component) as receiver.

```sh
make component
RUST_LOG=error wrpc-wasmtime-nats serve -n nats://127.0.0.1:4222 wasmtime wasmtime blaster-component/blaster-component.wasm
```

For the `init` context:

```javascript
// create a wrpc blaster client with nats transport
// each VU will have a dedicated nats connection
let blaster = wrpc.blaster({
  nats: {
    url: "nats://localhost:4222",
    prefix: "wasmtime",
  },
  // metric tags
  tags: { scenario: "basic" },
});
```

For the `scenario` context:

```javascript
// send a single packet (io blocking)
blaster.blast();

// tell component to burn cpu for 100ms
blaster.blast({
  cpu_burn_ms: 100,
});
```

`blaster`

- `blast(packet)`

`packet` is an object that can tell the wasm component to change behaviour:

- `cpu_burn_ms` (integer): Tell the component to burn cpu for X milliseconds
- `payload` (string): Arbitrary payload for the packet
- `memory_burn_mb` (integer): Tell the component to allocate X mb memory
- `wait_ms` (integer): Tell the component to sleep for X milliseconds
- `timeout` (integer): Request timeout in milliseconds

## HTTP API

For the `init` context:

```javascript
// create a wrpc http client with nats transport
// each VU will have a dedicated nats connection
let httpPacketGen = wrpc.http({
  nats: {
    url: "nats://localhost:4222",
    prefix: "default.AtVWn5-http_server",
  },
  // metric tags
  tags: { scenario: "basic" },
});
```

For the `scenario` context:

```javascript
// send a http get request (io blocking)
httpPacketGen.get("http://localhost:8000/");
```

`http`

Trying to bring as much as possible from [k6-http](https://grafana.com/docs/k6/latest/javascript-api/k6-http/).

- `expectedStatuses(statucCodes)`
- `get(url, [params])`
- `head(url, [params])`
- `patch(url, [body], [params])`
- `post(url, [body], [params])`
- `put(url, [body], [params])`
- `request(method, url, [body], [params])`
- `asyncRequest(method, url, [body], [params])`

`params` is an object like [k6-http/Params](https://grafana.com/docs/k6/latest/javascript-api/k6-http/params/) with:

- `auth`
- `cookies`
- `headers`
- `jar`
- `tags`
- `timeout`
