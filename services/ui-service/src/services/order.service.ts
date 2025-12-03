import type { CartItem } from "../types/product.types";
import { environment } from "../config/environment";

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
  private readonly baseURL = environment.ORDER_SERVICE_URL.startsWith('/')
    ? environment.ORDER_SERVICE_URL
    : `${environment.ORDER_SERVICE_URL}/order`;

  async createOrder(cartItems: CartItem[]): Promise<CreateOrderResponse> {
    const tokenString = localStorage.getItem("token");
    if (!tokenString) {
      throw new Error("Usuário não autenticado");
    }

    const token = JSON.parse(tokenString);

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
        Authorization: `Bearer ${token.accessToken}`,
      },
      body: JSON.stringify(items),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Erro ao criar pedido" }));
      throw new Error(error.error || "Erro ao criar pedido");
    }

    return response.json();
  }

  async getUserOrders(
    page: number = 1,
    pageSize: number = 10
  ): Promise<any> {
    const tokenString = localStorage.getItem("token");
    if (!tokenString) {
      throw new Error("Usuário não autenticado");
    }

    const token = JSON.parse(tokenString);
    const url = `${this.baseURL}/user/orders?page=${page}&pageSize=${pageSize}`;
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${token.accessToken}`,
      },
    });

    if (!response.ok) {
      throw new Error("Erro ao buscar pedidos do usuário");
    }

    return response.json();
  }

  async getUserOrderById(orderId: string): Promise<Order> {
    const tokenString = localStorage.getItem("token");
    if (!tokenString) {
      throw new Error("Usuário não autenticado");
    }

    const token = JSON.parse(tokenString);
    const url = `${this.baseURL}/user/order?id=${orderId}`;
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${token.accessToken}`,
      },
    });

    if (!response.ok) {
      throw new Error("Erro ao buscar pedido");
    }

    return response.json();
  }

  createOrderStatusStream(orderId: string): EventSource {
    const tokenString = localStorage.getItem("token");
    if (!tokenString) {
      throw new Error("Usuário não autenticado");
    }

    const token = JSON.parse(tokenString);
    const url = `${this.baseURL}/user/order/status?id=${orderId}&token=${encodeURIComponent(token.accessToken)}`;
    
    return new EventSource(url);
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

  async getOrdersByPage(
    page: number,
    pageSize: number
  ): Promise<{
    orders: Order[];
    totalCount: number;
    page: number;
    pageSize: number;
    totalPages: number;
  }> {
    const url = `${this.baseURL}/orders?page=${page}&pageSize=${pageSize}`;
    const response = await fetch(url);

    if (!response.ok) {
      throw new Error("Erro ao buscar pedidos paginados");
    }

    return response.json();
  }
}

export interface Order {
  id: string;
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
  updated_at: string;
}

export const orderService = new OrderService();
