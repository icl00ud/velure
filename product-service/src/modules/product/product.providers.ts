import { Connection } from 'mongoose';
import { ProductSchema } from './schemas/product.schema';
import { DATABASE_CONNECTION, PRODUCT_MODEL } from '../../shared/constants';

export const productsProviders = [
  {
    provide: PRODUCT_MODEL,
    useFactory: (connection: Connection) => connection.model('Product', ProductSchema),
    inject: [DATABASE_CONNECTION],
  },
];
