import { useEffect, useState } from "react";
import { productService } from "../services/product.service";
import type { Product } from "../types/product.types";

export function useProducts() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const fetchProducts = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await productService.getProducts();
      setProducts(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro ao buscar produtos");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProducts();
  }, []);

  return {
    products,
    loading,
    error,
    refetch: fetchProducts,
  };
}

export function useProduct(id: number) {
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const fetchProduct = async () => {
    if (!id) return;

    setLoading(true);
    setError(null);
    try {
      const data = await productService.getProductById(id);
      setProduct(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro ao buscar produto");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProduct();
  }, [id]);

  return {
    product,
    loading,
    error,
    refetch: fetchProduct,
  };
}

export function useProductsPaginated(page: number = 1, pageSize: number = 10, category?: string) {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState<number>(0);
  const [totalPages, setTotalPages] = useState<number>(0);

  const fetchProducts = async () => {
    setLoading(true);
    setError(null);
    try {
      let response;
      if (category) {
        response = await productService.getProductsByPageAndCategory(page, pageSize, category);
      } else {
        response = await productService.getProductsByPage(page, pageSize);
      }
      console.log("API Response:", response);
      setProducts(response.products || []);
      setTotalCount(response.totalCount || 0);
      setTotalPages(response.totalPages || 0);
    } catch (err) {
      console.error("Error fetching products:", err);
      setError(err instanceof Error ? err.message : "Erro ao buscar produtos");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProducts();
  }, [page, pageSize, category]);

  return {
    products,
    loading,
    error,
    totalCount,
    totalPages,
    refetch: fetchProducts,
  };
}

export function useCategories() {
  const [categories, setCategories] = useState<string[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const fetchCategories = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await productService.getCategories();
      setCategories(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro ao buscar categorias");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCategories();
  }, []);

  return {
    categories,
    loading,
    error,
    refetch: fetchCategories,
  };
}
