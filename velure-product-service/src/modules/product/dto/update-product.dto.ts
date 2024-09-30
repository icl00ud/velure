import { IsNotEmpty, IsOptional, IsNumber, IsString } from 'class-validator';

export class UpdateProductDto {
    @IsOptional()
    @IsString()
    @IsNotEmpty()
    name?: string;

    @IsOptional()
    @IsNumber()
    price?: number;

    @IsOptional()
    @IsString()
    @IsNotEmpty()
    description?: string;
}