export interface Product {
    name: string;
    description?: string;
    price: number;
    category?: string;
    disponibility: boolean;
    quantity_warehouse: number;
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