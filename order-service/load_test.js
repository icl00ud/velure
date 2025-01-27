import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  vus: 5000,
  duration: "5m",
};

export default function () {
  const url = "http://localhost:8000/publish-order";
  const payload = JSON.stringify({
    id: "12345",
    product_id: "67890",
    quantity: 2,
    total_amount: 500,
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  let res = http.post(url, payload, params);

  check(res, {
    "status is 201": (r) => r.status === 201,
  });

  sleep(0.5);
}
