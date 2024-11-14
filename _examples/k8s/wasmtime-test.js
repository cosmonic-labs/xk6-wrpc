import wrpc from "k6/x/wrpc";

export const options = {
  scenarios: {
    blaster: {
      executor: "constant-vus",
      vus: 50,
      duration: "15m",
      exec: "wrpcBlaster",
    },
    // http: {
    //   executor: "constant-vus",
    //   vus: 10,
    //   duration: "15m",
    //   startTime: "15m",
    //   exec: "wrpcHttpBlaster",
    // },
  },
};

let blaster = wrpc.blaster({
  nats: {
    url: "nats://nats-headless.default.svc.cluster.local:4222",
    prefix: "wasmtime",
  },
});

let wrpcHttp = wrpc.http({
  nats: {
    url: "nats://nats-headless.default.svc.cluster.local:4222",
    prefix: "wasmtime",
  },
});

export function wrpcBlaster() {
  blaster.blast();
}

export function wrpcHttpBlaster() {
  wrpcHttp.get("http://localhost:8080/");
}
