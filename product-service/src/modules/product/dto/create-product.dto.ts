import { IsArray, IsBase64, IsBoolean, IsNotEmpty, IsNumber, IsString } from "class-validator";

export class CreateProductDto {
    @IsString()
    @IsNotEmpty()
    readonly name: string;

    @IsString()
    readonly description?: string;

    @IsNumber()
    @IsNotEmpty()
    readonly price: number;

    @IsString()
    readonly category?: string;
    
    @IsBoolean()
    @IsNotEmpty()
    readonly disponibility: boolean;
    
    @IsNumber()
    @IsNotEmpty()
    readonly quantity_warehouse: number;
    
    @IsArray()
    @IsBase64()
    readonly images: string[];

    @IsNumber({}, { each: true})
    readonly dimensions: {
      readonly height?: number;
      readonly width?: number;
      readonly length?: number;
      readonly weight?: number;
    };

    @IsString()
    readonly brand?: string;

    @IsArray()
    @IsString({ each: true })
    readonly colors: string[];
    
    @IsString()
    readonly sku?: string;
  }