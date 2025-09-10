import { useState, useEffect } from 'react';
import { productService } from '../services/product.service';
import { Product } from '../types/product.types';

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
      setError(err instanceof Error ? err.message : 'Erro ao buscar produtos');
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
      setError(err instanceof Error ? err.message : 'Erro ao buscar produto');
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

  const fetchProducts = async () => {
    setLoading(true);
    setError(null);
    try {
      let data: Product[];
      if (category) {
        data = await productService.getProductsByPageAndCategory(page, pageSize, category);
      } else {
        data = await productService.getProductsByPage(page, pageSize);
      }
      setProducts(data);
      
      // Buscar contagem total apenas na primeira vez ou quando a categoria muda
      if (page === 1) {
        const count = await productService.getProductsCount();
        setTotalCount(count);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao buscar produtos');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProducts();
  }, [page, pageSize, category]);

  const totalPages = Math.ceil(totalCount / pageSize);

  return {
    products,
    loading,
    error,
    totalCount,
    totalPages,
    refetch: fetchProducts,
  };
}
