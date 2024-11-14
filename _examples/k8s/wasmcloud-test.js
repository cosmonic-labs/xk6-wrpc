import wrpc from "k6/x/wrpc";

export const options = {
  scenarios: {
    blaster: {
      executor: "constant-vus",
      vus: 10,
      duration: "15m",
      exec: "wrpcBlaster",
    },
    // http: {
    //   executor: "constant-vus",
    //   vus: 10,
    //   duration: "15m",
    //   exec: "httpBlaster",
    // },
  },
};

let blaster = wrpc.blaster({
  nats: {
    url: "nats://nats-headless.default.svc.cluster.local:4222",
    prefix: "default.blaster_component-component",
  },
});

let http = wrpc.http({
  nats: {
    url: "nats://nats-headless.default.svc.cluster.local:4222",
    prefix: "default.rust_hello_world-http_component",
  },
});

export function wrpcBlaster() {
  blaster.blast();
}

export function httpBlaster() {
  http.get("http://localhost:8080/");
}
