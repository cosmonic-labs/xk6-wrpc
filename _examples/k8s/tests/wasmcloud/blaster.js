import wrpc from "k6/x/wrpc";

export const options = {
  scenarios: {
    blue: {
      executor: "constant-vus",
      vus: 50,
      duration: "5m",
      exec: "blastBlue",
    },
    green: {
      executor: "constant-vus",
      vus: 50,
      duration: "5m",
      //      startTime: "5m",
      exec: "blastGreen",
    },
    // both: {
    //   executor: "constant-vus",
    //   vus: 50,
    //   duration: "15m",
    //   startTime: "10m",
    //   exec: "blastBoth",
    // },
  },
};

let blueBlaster = wrpc.blaster({
  nats: {
    url: "nats://nats-headless:4222",
    prefix: "default.blaster-component_blue",
  },
  tags: {
    variant: "blue",
  },
});

let greenBlaster = wrpc.blaster({
  nats: {
    url: "nats://nats-headless:4222",
    prefix: "default.blaster-component_green",
  },
  tags: {
    variant: "green",
  },
});

export function blastBoth() {
  blastBlue();
  blastGreen();
}

export function blastBlue() {
  blueBlaster.blast();
}

export function blastGreen() {
  greenBlaster.blast();
}
