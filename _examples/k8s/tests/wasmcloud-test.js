import wrpc from "k6/x/wrpc";
import http from "k6/http";

export const options = {
  scenarios: {
    blaster: {
      executor: "constant-vus",
      vus: 100,
      duration: "15m",
      exec: "wrpcBlaster",
    },
    wrpchttp: {
      executor: "constant-vus",
      vus: 100,
      duration: "15m",
      startTime: "15m",
      exec: "wrpcHttpBlaster",
    },
    http: {
      executor: "constant-vus",
      vus: 100,
      duration: "15m",
      startTime: "30m",
      exec: "httpBlaster",
    },
  },
};

let blaster = wrpc.blaster({
  nats: {
    url: "nats://nats-headless:4222",
    prefix: "default.blaster_component-component",
  },
});

let wrpcHttp = wrpc.http({
  nats: {
    url: "nats://nats-headless:4222",
    prefix: "default.rust_hello_world-http_component",
  },
});

export function wrpcBlaster() {
  blaster.blast();
}

export function wrpcHttpBlaster() {
  wrpcHttp.get("http://localhost:8080/");
}

export function httpBlaster() {
  http.get("http://wasmcloud-http-headless:8080/");
}
