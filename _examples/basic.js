import wrpc from "k6/x/wrpc";
import http from "k6/http";

export const options = {
  scenarios: {
    contacts: {
      executor: "constant-vus",
      vus: 1,
      duration: "10m",
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

export default function () {
  httpPacketGen.get("http://localhost:8000/");
  //  httpPacketGen.get("http://localhost:8000/error");
  // http.get("http://localhost:8000/error");
}
