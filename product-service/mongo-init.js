function printLargeText(text, numLines) {
  for (let i = 0; i < numLines; i++) {
      console.log(text);
  }
}

var rootUsername = process.env.MONGO_INITDB_ROOT_USERNAME;
var rootPassword = process.env.MONGO_INITDB_ROOT_PASSWORD;

var conn = new Mongo();
var db = conn.getDB("admin");

db.auth(rootUsername, rootPassword);

var dbName = process.env.MONGO_INITDB_DATABASE;
var collectionName = "products";

db = conn.getDB(dbName);
db.createCollection(collectionName);

db.createUser(
  {
    user: rootUsername,
    pwd:  rootPassword,
    roles: [ { role: "readWrite", db: dbName } ]
  }
)

// Insert 20 fake products
db[collectionName].insertMany([
  { 
    name: "Product 1", 
    description: "Description of Product 1", 
    price: 10.99,
    rating: 2.8,
    category: "Category 1", 
    disponibility: true, 
    quantity_warehouse: 50, 
    images: [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    dimensions: {
      height: 10,
      width: 5,
      length: 15,
      weight: 1.5
    },
    brand: "Brand 1",
    colors: ["Red", "Blue", "Green"],
    sku: "SKU12345"
  },
  { 
    name: "Product 2", 
    description: "Description of Product 2", 
    price: 15.49,
    rating: 3.5,
    category: "Category 2", 
    disponibility: false, 
    quantity_warehouse: 20, 
    images: [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    dimensions: {
      height: 8,
      width: 6,
      length: 12,
      weight: 1.2
    },
    brand: "Brand 2",
    colors: ["Black", "White"],
    sku: "SKU67890"
  },
  { 
    name: "Product 3", 
    description: "Description of Product 3", 
    price: 29.99, 
    rating: 4.2,
    category: "Category 3", 
    disponibility: true, 
    quantity_warehouse: 120, 
    images: [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    dimensions: {
      height: 12,
      width: 8,
      length: 20,
      weight: 2.0
    },
    brand: "Brand 3",
    colors: ["Yellow", "Orange"],
    sku: "SKU54321"
  },
  { 
    name: "Product 4", 
    description: "Description of Product 4", 
    price: 39.99, 
    rating: 4.8,
    category: "Category 2", 
    disponibility: false, 
    quantity_warehouse: 10, 
    images: [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    dimensions: {
      height: 15,
      width: 10,
      length: 25,
      weight: 2.5
    },
    brand: "Brand 4",
    colors: ["Purple", "Brown"],
    sku: "SKU24680"
  },
  { 
    name: "Product 5", 
    description: "Description of Product 5", 
    price: 49.99, 
    rating: 3.9,
    category: "Category 1", 
    disponibility: true, 
    quantity_warehouse: 75, 
    images: [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    dimensions: {
      height: 18,
      width: 12,
      length: 30,
      weight: 3.0
    },
    brand: "Brand 5",
    colors: ["Gray", "Silver"],
    sku: "SKU97531"
  },
  { 
    name: "Product 6", 
    description: "Description of Product 6", 
    price: 19.99, 
    rating: 3.2,
    category: "Category 3", 
    disponibility: true, 
    quantity_warehouse: 30, 
    images: [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    dimensions: {
      height: 14,
      width: 9,
      length: 22,
      weight: 2.2
    },
    brand: "Brand 6",
    colors: ["Gold", "Copper"],
    sku: "SKU86420"
  },
  { 
    name: "Product 7", 
    description: "Description of Product 7", 
    price: 59.99, 
    rating: 4.1,
    category: "Category 2", 
    disponibility: false, 
    quantity_warehouse: 5, 
    images: [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    dimensions: {
      height: 16,
      width: 11,
      length: 28,
      weight: 2.8
    },
    brand: "Brand 7",
    colors: ["Indigo", "Magenta"],
    sku: "SKU75319"
  },
  { 
    name: "Product 8", 
    description: "Description of Product 8", 
    price: 79.99, 
    rating: 4.6,
    category: "Category 1", 
    disponibility: true, 
    quantity_warehouse: 90, 
    images: [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    dimensions: {
      height: 20,
      width: 13,
      length: 35,
      weight: 3.5
    },
    brand: "Brand 8",
    colors: ["Turquoise", "Teal"],
    sku: "SKU31467"
  },
  { 
    "name": "Product 9", 
    "description": "Description of Product 9", 
    "price": 129.99, 
    "rating": 4.3,
    "category": "Category 3", 
    "disponibility": true, 
    "quantity_warehouse": 25, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 25,
      "width": 18,
      "length": 40,
      "weight": 4.2
    },
    "brand": "Brand 9",
    "colors": ["Crimson", "Maroon"],
    "sku": "SKU97654"
  },
  { 
    "name": "Product 10", 
    "description": "Description of Product 10", 
    "price": 39.99, 
    "rating": 3.8,
    "category": "Category 2", 
    "disponibility": false, 
    "quantity_warehouse": 0, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 15,
      "width": 10,
      "length": 30,
      "weight": 2.5
    },
    "brand": "Brand 10",
    "colors": ["Slate Gray", "Dark Slate Gray"],
    "sku": "SKU85321"
  },
  { 
    "name": "Product 11", 
    "description": "Description of Product 11", 
    "price": 89.99, 
    "rating": 4.5,
    "category": "Category 4", 
    "disponibility": true, 
    "quantity_warehouse": 15, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 22,
      "width": 16,
      "length": 36,
      "weight": 3.9
    },
    "brand": "Brand 11",
    "colors": ["Olive", "Forest Green"],
    "sku": "SKU11111"
  },
  { 
    "name": "Product 12", 
    "description": "Description of Product 12", 
    "price": 49.99, 
    "rating": 4.0,
    "category": "Category 1", 
    "disponibility": true, 
    "quantity_warehouse": 50, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 18,
      "width": 12,
      "length": 32,
      "weight": 3.2
    },
    "brand": "Brand 12",
    "colors": ["Silver", "Steel Blue"],
    "sku": "SKU22222"
  },
  { 
    "name": "Product 13", 
    "description": "Description of Product 13", 
    "price": 69.99, 
    "rating": 4.8,
    "category": "Category 3", 
    "disponibility": true, 
    "quantity_warehouse": 20, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 19,
      "width": 14,
      "length": 34,
      "weight": 3.4
    },
    "brand": "Brand 13",
    "colors": ["Gold", "Bronze"],
    "sku": "SKU33333"
  },
  { 
    "name": "Product 14", 
    "description": "Description of Product 14", 
    "price": 99.99, 
    "rating": 4.7,
    "category": "Category 2", 
    "disponibility": true, 
    "quantity_warehouse": 30, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 21,
      "width": 15,
      "length": 38,
      "weight": 3.7
    },
    "brand": "Brand 14",
    "colors": ["Lime", "Emerald"],
    "sku": "SKU44444"
  },
  { 
    "name": "Product 15", 
    "description": "Description of Product 15", 
    "price": 149.99, 
    "rating": 4.9,
    "category": "Category 4", 
    "disponibility": true, 
    "quantity_warehouse": 10, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 24,
      "width": 17,
      "length": 42,
      "weight": 4.5
    },
    "brand": "Brand 15",
    "colors": ["Aqua", "Sky Blue"],
    "sku": "SKU55555"
  },
  { 
    "name": "Product 16", 
    "description": "Description of Product 16", 
    "price": 59.99, 
    "rating": 4.2,
    "category": "Category 1", 
    "disponibility": true, 
    "quantity_warehouse": 70, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 17,
      "width": 12,
      "length": 30,
      "weight": 3.0
    },
    "brand": "Brand 16",
    "colors": ["Beige", "Tan"],
    "sku": "SKU66666"
  },
  { 
    "name": "Product 17", 
    "description": "Description of Product 17", 
    "price": 45.99, 
    "rating": 4.0,
    "category": "Category 2", 
    "disponibility": true, 
    "quantity_warehouse": 40, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 16,
      "width": 10,
      "length": 30,
      "weight": 2.5
    },
    "brand": "Brand 17",
    "colors": ["Charcoal", "Slate"],
    "sku": "SKU77777"
  },
  { 
    "name": "Product 18", 
    "description": "Description of Product 18", 
    "price": 99.99, 
    "rating": 4.7,
    "category": "Category 3", 
    "disponibility": true, 
    "quantity_warehouse": 20, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 20,
      "width": 15,
      "length": 35,
      "weight": 3.2
    },
    "brand": "Brand 18",
    "colors": ["Coral", "Peach"],
    "sku": "SKU88888"
  },
  { 
    "name": "Product 19", 
    "description": "Description of Product 19", 
    "price": 79.99, 
    "rating": 4.4,
    "category": "Category 1", 
    "disponibility": true, 
    "quantity_warehouse": 15, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 18,
      "width": 12,
      "length": 32,
      "weight": 3.0
    },
    "brand": "Brand 19",
    "colors": ["Navy Blue", "Royal Blue"],
    "sku": "SKU99999"
  },
  { 
    "name": "Product 20", 
    "description": "Description of Product 20", 
    "price": 119.99, 
    "rating": 4.6,
    "category": "Category 2", 
    "disponibility": true, 
    "quantity_warehouse": 25, 
    "images": [
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale",
      "https://picsum.photos/160/120?grayscale"
    ], 
    "dimensions": {
      "height": 22,
      "width": 16,
      "length": 38,
      "weight": 3.5
    },
    "brand": "Brand 20",
    "colors": ["Amethyst", "Violet"],
    "sku": "SKU10101"
  }
]);

printLargeText("Database and collection created successfully!", 12);