# wRPC extension for k6 load testing tool

🚧 **This is a work in progress.** 🚧

```sh
make
./k6 run ./_examples/basic.js
```

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
  - [ ] Ignore Responses
  - [ ] Callback for Responses
  - [ ] Asynchronous Requests
  - [ ] Metrics

## API

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
