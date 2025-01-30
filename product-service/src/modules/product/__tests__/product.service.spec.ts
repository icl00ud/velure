import { Test, TestingModule } from '@nestjs/testing';
import { ProductService } from '../product.service';
import { ProductRepository } from '../product.repository';
import { CreateProductDto } from '../dto/create-product.dto';

describe('ProductService', () => {
  let service: ProductService;
  let repository: ProductRepository;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        ProductService,
        {
          provide: ProductRepository,
          useValue: {
            getAllProducts: jest.fn().mockResolvedValue([]),
            getProductsByName: jest.fn().mockResolvedValue([]),
            getProductsByPage: jest.fn().mockResolvedValue([]),
            getProductsByPageAndCategory: jest.fn().mockResolvedValue([]),
            getProductsCount: jest.fn().mockResolvedValue(0),
            createProduct: jest.fn().mockResolvedValue({}),
            deleteProductsByName: jest.fn().mockResolvedValue(undefined),
            deleteProductById: jest.fn().mockResolvedValue(undefined),
          },
        },
      ],
    }).compile();

    service = module.get<ProductService>(ProductService);
    repository = module.get<ProductRepository>(ProductRepository);
  });

  it('should be defined', () => {
    expect(service).toBeDefined();
  });

  it('should return all products', async () => {
    await expect(service.getAllProducts()).resolves.toEqual([]);
  });

  it('should return products by name', async () => {
    const name = 'test';
    await expect(service.getProductsByName(name)).resolves.toEqual([]);
  });

  it('should return products by page', async () => {
    const page = 1;
    const pageSize = 10;
    await expect(service.getProductsByPage(page, pageSize)).resolves.toEqual([]);
  });

  it('should return products count', async () => {
    await expect(service.getProductsCount()).resolves.toEqual(0);
  });

  it('should create a product', async () => {
    const createProductDto: CreateProductDto = {
      name: 'test',
      price: 10,
      disponibility: true,
      quantity_warehouse: 10,
      images: [],
      dimensions: {},
      colors: [],
    };
    await expect(service.createProduct(createProductDto)).resolves.toEqual({});
  });

  it('should delete a product by name', async () => {
    const name = 'test';
    await expect(service.deleteProductsByName(name)).resolves.toBeUndefined();
  });

  it('should delete a product by id', async () => {
    const id = 'test';
    await expect(service.deleteProductById(id)).resolves.toBeUndefined();
  });
});