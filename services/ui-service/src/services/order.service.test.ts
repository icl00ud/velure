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
  const originalEventSource = global.EventSource;

  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  afterAll(() => {
    global.EventSource = originalEventSource;
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

  it("creates an EventSource stream with the token", () => {
    localStorage.setItem("token", JSON.stringify(token));
    const eventSourceSpy = vi.fn();

    class MockEventSource {
      constructor(public url: string) {
        eventSourceSpy(url);
      }
      close() {}
    }

    global.EventSource = MockEventSource as any;

    orderService.createOrderStatusStream("order-1");

    expect(eventSourceSpy).toHaveBeenCalledWith(
      "/api/order/user/order/status?id=order-1&token=token-123",
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
