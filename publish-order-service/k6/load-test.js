import http from "k6/http";
import { sleep, check } from "k6";

export let options = {
  vus: 100,
  duration: "30s",
};

export default function () {
  const counter = __ITER + 1;
  const url = "http://localhost:8080/create-order";
  const payload = JSON.stringify({
    items: [
      {
        product_id: "67a2c1e5564dfbe318544ca7",
        name: "Product 1",
        quantity: counter,
        price: 10.99,
      },
      {
        product_id: "67a2c1e5564dfbe318544ca8",
        name: "Product 2",
        quantity: counter + 1,
        price: 15.49,
      },
    ],
  });

  const params = {
    headers: { "Content-Type": "application/json" },
  };

  const res = http.post(url, payload, params);
  check(res, {
    "status is 201": (r) => r.status === 201,
  });
  sleep(1);
}
