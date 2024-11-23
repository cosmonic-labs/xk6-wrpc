import http from "k6/http";

export const options = {
  scenarios: {
    test: {
      executor: "constant-vus",
      vus: 50,
      duration: "10m",
    },
  },
};

const body = JSON.stringify({});
const params = { headers: { "Content-Type": "application/json" } };
const url = "http://invoker-headless:8080/action-name";

export default function () {
  http.post(url, body, params);
}
