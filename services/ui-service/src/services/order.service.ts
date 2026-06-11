import { environment } from "../config/environment";
import type { CartItem } from "../types/product.types";

export interface CreateOrderRequest {
  items: Array<{
    product_id: string;
    name: string;
    quantity: number;
    price: number;
  }>;
}

export interface CreateOrderResponse {
  order_id: string;
  total: number;
  status: string;
}

class OrderService {
  private readonly apiBase = environment.ORDER_SERVICE_URL.replace(/\/orders?$/, "");

  // Authentication: the auth-service sets an httpOnly access_token cookie at
  // login, and same-origin fetches send it automatically — no Authorization
  // header or localStorage needed (and XSS cannot read the cookie).
  async createOrder(cartItems: CartItem[]): Promise<CreateOrderResponse> {
    const items = cartItems.map((item) => ({
      product_id: item.product._id,
      name: item.product.name,
      quantity: item.quantity,
      price: item.product.price,
    }));

    const response = await fetch(`${this.apiBase}/orders`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(items),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Failed to create order" }));
      throw new Error(error.error || "Failed to create order");
    }

    return response.json();
  }

  async getUserOrders(page: number = 1, pageSize: number = 10): Promise<any> {
    const url = `${this.apiBase}/me/orders?page=${page}&pageSize=${pageSize}`;
    const response = await fetch(url);

    if (!response.ok) {
      throw new Error("Failed to fetch user orders");
    }

    return response.json();
  }

  async getUserOrderById(orderId: string): Promise<Order> {
    const url = `${this.apiBase}/me/orders/${orderId}`;
    const response = await fetch(url);

    if (!response.ok) {
      throw new Error("Failed to fetch order");
    }

    return response.json();
  }

  createOrderStatusStream(
    orderId: string,
    onMessage: (order: Order) => void,
    onError?: (error: Event) => void
  ): () => void {
    const url = `${this.apiBase}/me/orders/${orderId}/events`;

    const abortController = new AbortController();

    const connectSSE = async () => {
      try {
        const response = await fetch(url, {
          method: "GET",
          headers: {
            Accept: "text/event-stream",
          },
          signal: abortController.signal,
        });

        if (!response.ok) {
          throw new Error("Failed to connect to the status stream");
        }

        const reader = response.body?.getReader();
        if (!reader) {
          throw new Error("Stream not supported");
        }

        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          buffer = lines.pop() || "";

          for (const line of lines) {
            if (line.startsWith("data: ")) {
              try {
                const data = JSON.parse(line.slice(6));
                onMessage(data);
              } catch {}
            }
          }
        }
      } catch (error) {
        if ((error as Error).name !== "AbortError") {
          onError?.(error as Event);
        }
      }
    };

    connectSSE();

    return () => abortController.abort();
  }
}

export interface Order {
  id: string;
  _id?: string;
  user_id: string;
  items: Array<{
    product_id: string;
    name: string;
    quantity: number;
    price: number;
  }>;
  total: number;
  status: string;
  created_at: string;
  createdAt?: string;
  updated_at: string;
  updatedAt?: string;
}

export const orderService = new OrderService();
