import { vi } from "vitest";
import type { CartItem, Product } from "../types/product.types";
import { orderService } from "./order.service";

// Authentication travels in the httpOnly access_token cookie, sent
// automatically on same-origin fetches — no Authorization header and no
// localStorage involved.

const product: Product = {
  _id: "p1",
  name: "Pet Food",
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
    vi.restoreAllMocks();
  });

  it("creates an order with formatted payload", async () => {
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

    expect(fetch).toHaveBeenCalledWith("/api/orders", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify([
        {
          product_id: "p1",
          name: "Pet Food",
          quantity: 2,
          price: 50,
        },
      ]),
    });
    expect(result).toEqual(response);
  });

  it("surfaces the server error when an unauthenticated order is rejected", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      json: vi.fn().mockResolvedValue({ error: "unauthorized" }),
    } as any);

    await expect(orderService.createOrder(cartItems)).rejects.toThrow("unauthorized");
  });

  it("fetches paginated user orders", async () => {
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

    expect(fetch).toHaveBeenCalledWith("/api/me/orders?page=2&pageSize=5");
    expect(result).toEqual(ordersResponse);
  });

  it("throws when fetching order details fails", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      json: vi.fn(),
    } as any);

    await expect(orderService.getUserOrderById("order-1")).rejects.toThrow("Failed to fetch order");
    expect(fetch).toHaveBeenCalledWith("/api/me/orders/order-1");
  });

  it("creates an SSE stream relying on the session cookie", () => {
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
      "/api/me/orders/order-1/events",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({
          Accept: "text/event-stream",
        }),
      })
    );
  });
});
