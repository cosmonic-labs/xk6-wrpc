import wrpc from "k6/x/wrpc";

export const options = {
  scenarios: {
    contacts: {
      executor: "constant-vus",
      vus: 1,
      duration: "1m",
    },
  },
};

export let blaster = wrpc.blaster({
  nats: {
    url: "nats://nats.default.svc.cluster.local:4222",
    prefix: "wasmtime",
  },
});

export function handleSummary(data) {
  const blaster_operation = data.metrics.wrpc_blaster_operation;
  const blaster_duration = data.metrics.wrpc_blaster_duration;
  let parts = [
    `vus: ${data.metrics.vus.values.value}`,
    `op_count: ${blaster_operation.values.count}`,
    `op_rate: ${blaster_operation.values.rate}`,
    `op_dur_avg: ${blaster_duration.values.avg}`,
    `op_dur_p95: ${blaster_duration.values["p(95)"]}`,
    `\n`,
  ];

  return {
    stdout: parts.join(" "),
  };
}

export default function () {
  blaster.blast();
}
