import * as mongoose from 'mongoose';

export const ProductSchema = new mongoose.Schema({
  name: { type: String, required: true },
  description: { type: String },
  price: { type: Number, required: true },
  category: { type: String },
  disponibility: { type: Boolean, default: true },
  quantity_warehouse: { type: Number, default: 0 },
  images: [{ type: String }],
  dimensions: {
    height: { type: Number },
    width: { type: Number },
    length: { type: Number },
    weight: { type: Number }
  },
  brand: { type: String },
  colors: [{ type: String }],
  sku: { type: String },
  dt_created: { type: Date, default: Date.now },
  dt_updated: { type: Date, default: Date.now }
});