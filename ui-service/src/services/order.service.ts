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
  private readonly baseURL = "http://localhost:3030";

  async createOrder(cartItems: CartItem[]): Promise<CreateOrderResponse> {
    const items = cartItems.map((item) => ({
      product_id: item.product._id,
      name: item.product.name,
      quantity: item.quantity,
      price: item.product.price,
    }));

    const response = await fetch(`${this.baseURL}/create-order`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(items),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Erro ao criar pedido" }));
      throw new Error(error.error || "Erro ao criar pedido");
    }

    return response.json();
  }

  async updateOrderStatus(orderId: string, status: string): Promise<void> {
    const response = await fetch(`${this.baseURL}/update-order-status`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        order_id: orderId,
        status: status,
      }),
    });

    if (!response.ok) {
      throw new Error("Erro ao atualizar status do pedido");
    }
  }
}

export const orderService = new OrderService();
