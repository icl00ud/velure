import { vi } from "vitest";
import type { Product } from "../types/product.types";
import { configService } from "./config.service";
import { productService } from "./product.service";

const mockProduct: Product = {
  _id: "1",
  name: "Coleira",
  description: "Coleira resistente",
  price: 25,
  rating: 4,
  quantity: 5,
  images: ["image.jpg"],
  dimensions: {},
  colors: [],
  dt_created: new Date(),
  dt_updated: new Date(),
};

describe("productService", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("fetches all products", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue([mockProduct]),
    } as any);

    const result = await productService.getProducts();

    expect(fetch).toHaveBeenCalledWith(configService.productServiceUrl);
    expect(result).toEqual([mockProduct]);
  });

  it("throws when fetching a product by id fails", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      json: vi.fn(),
    } as any);

    await expect(productService.getProductById(42)).rejects.toThrow("Erro ao buscar produto");
  });

  it("creates a new product with the correct payload", async () => {
    const creationPayload: Omit<Product, "_id" | "dt_created" | "dt_updated"> = {
      name: "Brinquedo",
      description: "Brinquedo interativo",
      price: 40,
      rating: 5,
      quantity: 12,
      images: ["toy.jpg"],
      dimensions: {},
      colors: [],
      brand: "Velure",
    };

    const createdProduct: Product = {
      ...creationPayload,
      _id: "99",
      dt_created: new Date(),
      dt_updated: new Date(),
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue(createdProduct),
    } as any);

    const result = await productService.createProduct(creationPayload);

    expect(fetch).toHaveBeenCalledWith(configService.productServiceUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(creationPayload),
    });
    expect(result).toEqual(createdProduct);
  });
});
