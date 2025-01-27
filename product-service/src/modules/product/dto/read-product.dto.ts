export class ReadProductDTO {
    readonly name: string;
    readonly description?: string;
    readonly price: number;
    readonly category?: string;
    readonly disponibility: boolean;
    readonly quantity_warehouse: number;
    readonly images: string[];
    readonly dimensions: {
      readonly height?: number;
      readonly width?: number;
      readonly length?: number;
      readonly weight?: number;
    };
    readonly brand?: string;
    readonly colors: string[];
    readonly sku?: string;
    readonly dt_created: Date;
    readonly dt_updated: Date;
  }  