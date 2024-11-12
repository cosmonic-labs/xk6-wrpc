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

let http = wrpc.http({
  nats: {
    url: "nats://localhost:4222",
    prefix: "default.AtVWn5-http_server",
  },
  tags: { scenario: "contacts" },
});

// NOTE(lxf): simple ascii conversion, not suitable for utf-8/other encodings
function bytesToString(buf) {
  return String.fromCharCode.apply(null, buf);
}

export default function () {
  // simple get
  http.get("http://localhost:8000/");

  // get returning body
  let resp = http.get("http://localhost:8000/", { consume: true });
  // the response object has the following fields:
  // - status: number
  // - headers: object
  // - body: Uint8Array
  // to convert the body to a string:
  // console.log(bytesToString(resp.body));

  // all options
  http.get("http://localhost:8000/", {
    // request timeout in ms
    timeout: 10000,
    // consume response body
    consume: true,
    // http basic auth
    auth: {
      username: "user",
      password: "pass",
    },
    // request headers
    headers: {
      "X-Header": "value",
    },
  });

  // post with json body
  http.post("http://localhost:8000/post", { hello: "world" });

  // post with json body, returning body
  resp = http.post(
    "http://localhost:8000/post",
    { hello: "world" },
    { consume: true },
  );
}
