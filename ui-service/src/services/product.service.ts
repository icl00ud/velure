import type { Product } from "../types/product.types";
import { configService } from "./config.service";

class ProductService {
  async getProducts(): Promise<Product[]> {
    try {
      const response = await fetch(configService.productServiceUrl);
      if (!response.ok) {
        throw new Error("Erro ao buscar produtos");
      }
      return await response.json();
    } catch (error) {
      console.error("Erro ao buscar produtos:", error);
      throw error;
    }
  }

  async getProductById(id: number): Promise<Product> {
    try {
      const response = await fetch(`${configService.productServiceUrl}/${id}`);
      if (!response.ok) {
        throw new Error("Erro ao buscar produto");
      }
      return await response.json();
    } catch (error) {
      console.error("Erro ao buscar produto:", error);
      throw error;
    }
  }

  async getProductsByPage(page: number, pageSize: number): Promise<Product[]> {
    try {
      const url = `${configService.productServiceUrl}/getProductsByPage?page=${page}&pageSize=${pageSize}`;
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error("Erro ao buscar produtos paginados");
      }
      return await response.json();
    } catch (error) {
      console.error("Erro ao buscar produtos paginados:", error);
      throw error;
    }
  }

  async getProductsByPageAndCategory(
    page: number,
    pageSize: number,
    productCategory: string
  ): Promise<Product[]> {
    try {
      const url = `${configService.productServiceUrl}/getProductsByPageAndCategory?page=${page}&pageSize=${pageSize}&category=${productCategory}`;
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error("Erro ao buscar produtos por categoria");
      }
      return await response.json();
    } catch (error) {
      console.error("Erro ao buscar produtos por categoria:", error);
      throw error;
    }
  }

  async getCategories(): Promise<string[]> {
    try {
      const url = `${configService.productServiceUrl}/categories`;
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error("Erro ao buscar categorias");
      }
      return await response.json();
    } catch (error) {
      console.error("Erro ao buscar categorias:", error);
      throw error;
    }
  }

  async getProductsCount(): Promise<number> {
    try {
      const url = `${configService.productServiceUrl}/getProductsCount`;
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error("Erro ao buscar contagem de produtos");
      }
      const data = await response.json();
      return data.count || 0;
    } catch (error) {
      console.error("Erro ao buscar contagem de produtos:", error);
      throw error;
    }
  }

  async createProduct(
    product: Omit<Product, "_id" | "dt_created" | "dt_updated">
  ): Promise<Product> {
    try {
      const response = await fetch(configService.productServiceUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(product),
      });

      if (!response.ok) {
        throw new Error("Erro ao criar produto");
      }
      return await response.json();
    } catch (error) {
      console.error("Erro ao criar produto:", error);
      throw error;
    }
  }

  async updateProduct(id: number, product: Partial<Product>): Promise<Product> {
    try {
      const response = await fetch(`${configService.productServiceUrl}/${id}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(product),
      });

      if (!response.ok) {
        throw new Error("Erro ao atualizar produto");
      }
      return await response.json();
    } catch (error) {
      console.error("Erro ao atualizar produto:", error);
      throw error;
    }
  }

  async deleteProduct(id: number): Promise<void> {
    try {
      const response = await fetch(`${configService.productServiceUrl}/${id}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        throw new Error("Erro ao deletar produto");
      }
    } catch (error) {
      console.error("Erro ao deletar produto:", error);
      throw error;
    }
  }
}

export const productService = new ProductService();
