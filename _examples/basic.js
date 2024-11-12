import wrpc from "k6/x/wrpc";
import http from "k6/http";

export const options = {
  scenarios: {
    contacts: {
      executor: "constant-vus",
      vus: 1,
      duration: "10s",
    },
  },
};

let httpPacketGen = wrpc.http({
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
  // http.post(
  //   "http://localhost:8000/post",
  //   { hello: "world" },
  //   {
  //     auth: {
  //       username: "admin",
  //       password: "admin",
  //     },
  //     headers: {
  //       "X-Custom-Header": "k6",
  //     },
  //     timeout: 10000,
  //     consume: true,
  //   },
  // );

  httpPacketGen.get("http://localhost:8000/");
}
