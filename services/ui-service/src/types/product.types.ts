export interface Product {
  _id: string;
  name: string;
  description?: string;
  price: number;
  rating: number;
  category?: string;
  quantity: number;
  images: string[];
  dimensions: {
    height?: number;
    width?: number;
    length?: number;
    weight?: number;
  };
  brand?: string;
  colors: string[];
  sku?: string;
  dt_created: Date;
  dt_updated: Date;
}

export interface CartItem {
  product: Product;
  quantity: number;
}
