import { renderHook, waitFor } from "@testing-library/react";
import { vi } from "vitest";
import {
  useCategories,
  useProduct,
  useProducts,
  useProductsPaginated,
} from "./use-products";
import { productService } from "../services/product.service";

vi.mock("../services/product.service", () => ({
  productService: {
    getProducts: vi.fn(),
    getProductById: vi.fn(),
    getProductsByPage: vi.fn(),
    getProductsByPageAndCategory: vi.fn(),
    getCategories: vi.fn(),
  },
}));

const mockedService = vi.mocked(productService);

const mockProduct = {
  _id: "1",
  name: "Ração Premium",
  price: 99.9,
  rating: 4.5,
  quantity: 10,
  images: ["image.jpg"],
  dimensions: {},
  colors: [],
  dt_created: new Date(),
  dt_updated: new Date(),
};

describe("use-products hooks", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("loads products on mount", async () => {
    mockedService.getProducts.mockResolvedValueOnce([mockProduct]);

    const { result } = renderHook(() => useProducts());

    expect(result.current.loading).toBe(true);

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(mockedService.getProducts).toHaveBeenCalled();
    expect(result.current.products).toEqual([mockProduct]);
    expect(result.current.error).toBeNull();
  });

  it("returns an error when fetching a product fails", async () => {
    mockedService.getProductById.mockRejectedValueOnce(new Error("Erro ao buscar produto"));

    const { result } = renderHook(() => useProduct(123));

    await waitFor(() => expect(result.current.error).toBe("Erro ao buscar produto"));
    expect(result.current.product).toBeNull();
    expect(result.current.loading).toBe(false);
  });

  it("fetches paginated products by category", async () => {
    mockedService.getProductsByPageAndCategory.mockResolvedValueOnce({
      products: [mockProduct],
      totalCount: 5,
      totalPages: 1,
      page: 1,
      pageSize: 10,
    });

    const { result } = renderHook(() => useProductsPaginated(1, 10, "dogs"));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(mockedService.getProductsByPageAndCategory).toHaveBeenCalledWith(1, 10, "dogs");
    expect(result.current.products).toEqual([mockProduct]);
    expect(result.current.totalCount).toBe(5);
    expect(result.current.totalPages).toBe(1);
    expect(result.current.error).toBeNull();
  });

  it("loads categories on mount", async () => {
    mockedService.getCategories.mockResolvedValueOnce(["dogs", "cats"]);

    const { result } = renderHook(() => useCategories());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(mockedService.getCategories).toHaveBeenCalled();
    expect(result.current.categories).toEqual(["dogs", "cats"]);
    expect(result.current.error).toBeNull();
  });
});
