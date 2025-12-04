import { vi } from "vitest";
import { orderService } from "./order.service";
import type { CartItem, Product } from "../types/product.types";

const token = { accessToken: "token-123" };

const product: Product = {
  _id: "p1",
  name: "Ração",
  price: 50,
  rating: 4.5,
  quantity: 10,
  images: ["img.jpg"],
  dimensions: {},
  colors: [],
  dt_created: new Date(),
  dt_updated: new Date(),
};

const cartItems: CartItem[] = [
  {
    product,
    quantity: 2,
  },
];

describe("orderService", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it("creates an order with token and formatted payload", async () => {
    localStorage.setItem("token", JSON.stringify(token));

    const response = {
      order_id: "order-1",
      total: 100,
      status: "pending",
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue(response),
    } as any);

    const result = await orderService.createOrder(cartItems);

    expect(fetch).toHaveBeenCalledWith("/api/order/create-order", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token.accessToken}`,
      },
      body: JSON.stringify([
        {
          product_id: "p1",
          name: "Ração",
          quantity: 2,
          price: 50,
        },
      ]),
    });
    expect(result).toEqual(response);
  });

  it("throws when creating an order without an auth token", async () => {
    await expect(orderService.createOrder(cartItems)).rejects.toThrow("Usuário não autenticado");
  });

  it("fetches paginated user orders", async () => {
    localStorage.setItem("token", JSON.stringify(token));

    const ordersResponse = {
      orders: [],
      totalCount: 0,
      page: 2,
      pageSize: 5,
      totalPages: 1,
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue(ordersResponse),
    } as any);

    const result = await orderService.getUserOrders(2, 5);

    expect(fetch).toHaveBeenCalledWith("/api/order/user/orders?page=2&pageSize=5", {
      headers: {
        Authorization: `Bearer ${token.accessToken}`,
      },
    });
    expect(result).toEqual(ordersResponse);
  });

  it("throws when fetching order details fails", async () => {
    localStorage.setItem("token", JSON.stringify(token));
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      json: vi.fn(),
    } as any);

    await expect(orderService.getUserOrderById("order-1")).rejects.toThrow("Erro ao buscar pedido");
  });

  it("creates an SSE stream using fetch with auth header", () => {
    localStorage.setItem("token", JSON.stringify(token));

    const mockReader = {
      read: vi.fn().mockResolvedValue({ done: true, value: undefined }),
    };
    const mockBody = {
      getReader: () => mockReader,
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      body: mockBody,
    } as any);

    const onMessage = vi.fn();
    orderService.createOrderStatusStream("order-1", onMessage);

    expect(fetch).toHaveBeenCalledWith(
      "/api/order/user/order/status?id=order-1",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({
          Authorization: `Bearer ${token.accessToken}`,
          Accept: "text/event-stream",
        }),
      }),
    );
  });

  it("throws when updating order status fails", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      json: vi.fn(),
    } as any);

    await expect(orderService.updateOrderStatus("order-1", "shipped")).rejects.toThrow(
      "Erro ao atualizar status do pedido",
    );
  });

  it("retrieves orders by page", async () => {
    const pagedOrders = {
      orders: [{ id: "order-1" }],
      totalCount: 1,
      page: 1,
      pageSize: 10,
      totalPages: 1,
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue(pagedOrders),
    } as any);

    const result = await orderService.getOrdersByPage(1, 10);

    expect(fetch).toHaveBeenCalledWith("/api/order/orders?page=1&pageSize=10");
    expect(result).toEqual(pagedOrders);
  });
});
