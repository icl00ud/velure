/**
 * Script de inicialização do MongoDB com produtos realistas de petshop
 * Usa dados reais e imagens de APIs gratuitas
 */

// Incluir os scripts de geração de produtos (simulação de require para MongoDB)
// Em ambiente de produção, você pode usar require() se suportado
load('scripts/generate-realistic-products.js');
load('scripts/pet-image-service.js');

function printLargeText(text, numLines) {
  for (let i = 0; i < numLines; i++) {
    print(text);
  }
}

var rootUsername = "root";
var rootPassword = "root";

var conn = new Mongo();
var adminDB = conn.getDB("admin");

// Autentica no banco de administração
adminDB.auth(rootUsername, rootPassword);

var dbName = "velure_database";
var collectionName = "products";

var db = conn.getDB(dbName);

// Cria a collection se não existir
db.createCollection(collectionName);

// Cria o usuário com permissão de readWrite no banco de dados especificado
db.createUser({
  user: rootUsername,
  pwd: rootPassword,
  roles: [{ role: "readWrite", db: dbName }],
});

// Produtos realistas de petshop
const realisticProducts = [
  // === PRODUTOS PARA CÃES ===
  {
    name: "Ração Premium Cães Adultos Frango e Arroz 15kg",
    description: "Ração super premium para cães adultos com frango real e arroz integral. Rico em proteínas de alta qualidade, glucosamina e condroitina para articulações saudáveis. Sem corantes artificiais.",
    price: 189.90,
    rating: 4.7,
    category: "Alimentação",
    disponibility: true,
    quantity_warehouse: 45,
    images: [
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop&q=80&seed=1",
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop&q=80&seed=2"
    ],
    dimensions: { height: 60, width: 40, length: 15, weight: 15.0 },
    brand: "Royal Canin",
    colors: ["Natural", "Marrom"],
    sku: "RC-DOG-ADULT-15KG",
  },
  {
    name: "Bola Interativa Kong com Dispenser de Petiscos",
    description: "Bola de borracha natural resistente com compartimento interno para petiscos. Estimula o exercício mental e físico. Ideal para cães de médio e grande porte.",
    price: 67.90,
    rating: 4.8,
    category: "Brinquedos",
    disponibility: true,
    quantity_warehouse: 32,
    images: [
      "https://images.unsplash.com/photo-1601758228041-f3b2795255f1?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1601758228041-f3b2795255f1?w=400&h=300&fit=crop&q=80&seed=3",
      "https://images.unsplash.com/photo-1601758228041-f3b2795255f1?w=400&h=300&fit=crop&q=80&seed=4"
    ],
    dimensions: { height: 12, width: 12, length: 12, weight: 0.4 },
    brand: "Kong",
    colors: ["Vermelho", "Azul", "Verde"],
    sku: "KONG-BALL-TREAT-M",
  },
  {
    name: "Coleira Ajustável Couro Legítimo com Fivela Inox",
    description: "Coleira premium de couro legítimo italiano com fivela de aço inoxidável. Acolchoada internamente para máximo conforto. Disponível em vários tamanhos.",
    price: 89.50,
    rating: 4.6,
    category: "Acessórios",
    disponibility: true,
    quantity_warehouse: 28,
    images: [
      "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=400&h=300&fit=crop&q=80&seed=5",
      "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=400&h=300&fit=crop&q=80&seed=6"
    ],
    dimensions: { height: 3, width: 45, length: 2, weight: 0.3 },
    brand: "LeatherPet",
    colors: ["Marrom", "Preto", "Caramelo"],
    sku: "LP-COLLAR-LEATHER-M",
  },
  {
    name: "Cama Ortopédica Memory Foam Grande",
    description: "Cama ortopédica com espuma memory foam de alta densidade. Capa removível e lavável, tecido antialérgico. Ideal para cães idosos ou com problemas articulares.",
    price: 299.90,
    rating: 4.9,
    category: "Camas e Descanso",
    disponibility: true,
    quantity_warehouse: 15,
    images: [
      "https://images.unsplash.com/photo-1558617047-ac1a6b5abbd7?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1558617047-ac1a6b5abbd7?w=400&h=300&fit=crop&q=80&seed=7",
      "https://images.unsplash.com/photo-1558617047-ac1a6b5abbd7?w=400&h=300&fit=crop&q=80&seed=8"
    ],
    dimensions: { height: 15, width: 80, length: 60, weight: 3.5 },
    brand: "ComfortPet",
    colors: ["Cinza", "Bege", "Azul"],
    sku: "CP-BED-ORTHO-L",
  },
  {
    name: "Petisco Natural Osso de Couro Bovino",
    description: "Osso de couro 100% natural bovino, sem aditivos químicos. Ideal para higiene dental e entretenimento. Longa duração, rico em colágeno.",
    price: 24.90,
    rating: 4.4,
    category: "Petiscos",
    disponibility: true,
    quantity_warehouse: 67,
    images: [
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop&q=80&seed=9",
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop&q=80&seed=10",
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop&q=80&seed=11"
    ],
    dimensions: { height: 15, width: 5, length: 20, weight: 0.2 },
    brand: "DogChew Natural",
    colors: ["Natural", "Bege"],
    sku: "DCN-BONE-NATURAL-L",
  },

  // === PRODUTOS PARA GATOS ===
  {
    name: "Ração Premium Gatos Castrados Frango 7.5kg",
    description: "Ração especial para gatos castrados com controle de peso. Baixo teor de gordura, rico em proteínas e fibras. Com cranberry para saúde urinária.",
    price: 145.90,
    rating: 4.6,
    category: "Alimentação",
    disponibility: true,
    quantity_warehouse: 38,
    images: [
      "https://images.unsplash.com/photo-1571566882372-1598d88abd90?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1571566882372-1598d88abd90?w=400&h=300&fit=crop&q=80&seed=12",
      "https://images.unsplash.com/photo-1571566882372-1598d88abd90?w=400&h=300&fit=crop&q=80&seed=13"
    ],
    dimensions: { height: 45, width: 30, length: 12, weight: 7.5 },
    brand: "Hill's Science Diet",
    colors: ["Natural"],
    sku: "HILLS-CAT-CAST-7.5KG",
  },
  {
    name: "Arranhador Torre com Plataformas e Brinquedos",
    description: "Torre arranhadora de sisal natural com múltiplas plataformas, esconderijos e brinquedos suspensos. Base estável e resistente.",
    price: 299.90,
    rating: 4.8,
    category: "Brinquedos",
    disponibility: true,
    quantity_warehouse: 12,
    images: [
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop&q=80&seed=14",
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop&q=80&seed=15"
    ],
    dimensions: { height: 120, width: 60, length: 40, weight: 18.0 },
    brand: "CatTree Premium",
    colors: ["Bege", "Cinza"],
    sku: "CTP-TOWER-120CM",
  },
  {
    name: "Caixa de Areia Fechada com Filtro Anti-Odor",
    description: "Caixa de areia fechada com sistema de filtro de carvão ativado. Porta basculante, fácil limpeza. Ideal para controle de odores.",
    price: 189.90,
    rating: 4.5,
    category: "Higiene",
    disponibility: true,
    quantity_warehouse: 22,
    images: [
      "https://placehold.co/400x300/4169E1/FFFFFF?text=Cat+Litter+Box",
      "https://placehold.co/400x300/4169E1/FFFFFF?text=Cat+Litter+Box&seed=1",
      "https://placehold.co/400x300/4169E1/FFFFFF?text=Cat+Litter+Box&seed=2"
    ],
    dimensions: { height: 40, width: 55, length: 40, weight: 2.8 },
    brand: "CleanCat Pro",
    colors: ["Branco", "Cinza", "Azul"],
    sku: "CCP-LITTER-CLOSED",
  },
  {
    name: "Sachê Gourmet Peixe ao Molho 85g - Pack 12un",
    description: "Alimento úmido gourmet com pedaços de peixe em molho saboroso. Rico em ômega 3 e nutrientes essenciais. Sem conservantes artificiais.",
    price: 42.90,
    rating: 4.7,
    category: "Alimentação",
    disponibility: true,
    quantity_warehouse: 89,
    images: [
      "https://images.unsplash.com/photo-1571566882372-1598d88abd90?w=400&h=300&fit=crop&q=80&seed=16",
      "https://images.unsplash.com/photo-1571566882372-1598d88abd90?w=400&h=300&fit=crop&q=80&seed=17",
      "https://images.unsplash.com/photo-1571566882372-1598d88abd90?w=400&h=300&fit=crop&q=80&seed=18"
    ],
    dimensions: { height: 12, width: 25, length: 18, weight: 1.02 },
    brand: "Whiskas",
    colors: ["Variadas"],
    sku: "WK-SACHET-FISH-12PK",
  },
  {
    name: "Varinha Interativa com Penas Naturais",
    description: "Brinquedo interativo com penas naturais coloridas e guizo. Estimula o instinto de caça dos felinos. Cabo telescópico ajustável.",
    price: 29.90,
    rating: 4.3,
    category: "Brinquedos",
    disponibility: true,
    quantity_warehouse: 54,
    images: [
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop&q=80&seed=19",
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop&q=80&seed=20",
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop&q=80&seed=21"
    ],
    dimensions: { height: 50, width: 5, length: 5, weight: 0.15 },
    brand: "FelinePlay",
    colors: ["Multicolor"],
    sku: "FP-WAND-FEATHER",
  },

  // === PRODUTOS PARA PÁSSAROS ===
  {
    name: "Mistura de Sementes Premium Canários 1kg",
    description: "Mistura especial de sementes selecionadas para canários. Rica em alpiste, níger e sementes nutritivas. Com vitaminas A, D3 e E.",
    price: 34.90,
    rating: 4.5,
    category: "Alimentação",
    disponibility: true,
    quantity_warehouse: 76,
    images: [
      "https://images.unsplash.com/photo-1598300042247-d088f8ab3a91?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1598300042247-d088f8ab3a91?w=400&h=300&fit=crop&q=80&seed=22",
      "https://images.unsplash.com/photo-1598300042247-d088f8ab3a91?w=400&h=300&fit=crop&q=80&seed=23"
    ],
    dimensions: { height: 25, width: 18, length: 8, weight: 1.0 },
    brand: "BirdSeed Premium",
    colors: ["Natural"],
    sku: "BSP-CANARY-MIX-1KG",
  },
  {
    name: "Gaiola Espaçosa com Poleiros Naturais",
    description: "Gaiola espaçosa com poleiros de madeira natural e comedouros em aço inox. Bandeja removível para fácil limpeza. Design moderno.",
    price: 449.90,
    rating: 4.7,
    category: "Gaiolas",
    disponibility: true,
    quantity_warehouse: 8,
    images: [
      "https://images.unsplash.com/photo-1555169062-013468b47731?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1555169062-013468b47731?w=400&h=300&fit=crop&q=80&seed=24",
      "https://images.unsplash.com/photo-1555169062-013468b47731?w=400&h=300&fit=crop&q=80&seed=25"
    ],
    dimensions: { height: 60, width: 45, length: 35, weight: 8.5 },
    brand: "BirdHome Deluxe",
    colors: ["Branco", "Preto"],
    sku: "BHD-CAGE-LARGE",
  },
  {
    name: "Suplemento Vitamínico para Pássaros",
    description: "Complexo vitamínico completo para aves. Fortalece o sistema imunológico, melhora a plumagem e aumenta a vitalidade.",
    price: 28.90,
    rating: 4.4,
    category: "Suplementos",
    disponibility: true,
    quantity_warehouse: 45,
    images: [
      "https://placehold.co/400x300/228B22/FFFFFF?text=Bird+Vitamins",
      "https://placehold.co/400x300/228B22/FFFFFF?text=Bird+Vitamins&seed=1",
      "https://placehold.co/400x300/228B22/FFFFFF?text=Bird+Vitamins&seed=2"
    ],
    dimensions: { height: 12, width: 8, length: 5, weight: 0.15 },
    brand: "VitaBird",
    colors: ["Transparente"],
    sku: "VB-VITAMINS-100ML",
  },

  // === PRODUTOS PARA PEIXES ===
  {
    name: "Ração em Flocos Peixes Tropicais 200g",
    description: "Ração balanceada em flocos para peixes tropicais. Rica em proteínas, vitaminas e minerais. Realça cores naturais dos peixes.",
    price: 24.90,
    rating: 4.6,
    category: "Alimentação",
    disponibility: true,
    quantity_warehouse: 98,
    images: [
      "https://images.unsplash.com/photo-1559827260-dc66d52bef19?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1559827260-dc66d52bef19?w=400&h=300&fit=crop&q=80&seed=26",
      "https://images.unsplash.com/photo-1559827260-dc66d52bef19?w=400&h=300&fit=crop&q=80&seed=27"
    ],
    dimensions: { height: 15, width: 10, length: 6, weight: 0.2 },
    brand: "AquaFood Premium",
    colors: ["Natural"],
    sku: "AFP-FLAKES-TROPICAL-200G",
  },
  {
    name: "Filtro Submerso Aquário 100L com Bomba",
    description: "Filtro submerso silencioso com bomba integrada. Sistema de filtragem biológica e mecânica. Ideal para aquários de 50 a 100 litros.",
    price: 129.90,
    rating: 4.5,
    category: "Aquários",
    disponibility: true,
    quantity_warehouse: 18,
    images: [
      "https://images.unsplash.com/photo-1544551763-46a013bb70d5?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1544551763-46a013bb70d5?w=400&h=300&fit=crop&q=80&seed=28",
      "https://images.unsplash.com/photo-1544551763-46a013bb70d5?w=400&h=300&fit=crop&q=80&seed=29"
    ],
    dimensions: { height: 25, width: 15, length: 10, weight: 1.8 },
    brand: "AquaTech Pro",
    colors: ["Preto"],
    sku: "ATP-FILTER-SUB-100L",
  },
  {
    name: "Aquário Completo LED 60L com Filtro",
    description: "Aquário completo com iluminação LED, filtro interno e termômetro. Inclui substrato e plantas artificiais. Kit completo para iniciantes.",
    price: 389.90,
    rating: 4.8,
    category: "Aquários",
    disponibility: true,
    quantity_warehouse: 6,
    images: [
      "https://images.unsplash.com/photo-1554263897-4bfa012dcac0?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1554263897-4bfa012dcac0?w=400&h=300&fit=crop&q=80&seed=30",
      "https://images.unsplash.com/photo-1554263897-4bfa012dcac0?w=400&h=300&fit=crop&q=80&seed=31"
    ],
    dimensions: { height: 35, width: 60, length: 30, weight: 12.0 },
    brand: "AquaStart Complete",
    colors: ["Transparente"],
    sku: "ASC-AQUARIUM-60L-KIT",
  },

  // === PRODUTOS GERAIS PARA PETS ===
  {
    name: "Shampoo Neutro para Todos os Pets 500ml",
    description: "Shampoo hipoalergênico para cães, gatos e outros pets. PH neutro, livre de sulfatos e parabenos. Com aloe vera e vitamina E.",
    price: 39.90,
    rating: 4.3,
    category: "Higiene",
    disponibility: true,
    quantity_warehouse: 67,
    images: [
      "https://placehold.co/400x300/9370DB/FFFFFF?text=Pet+Shampoo",
      "https://placehold.co/400x300/9370DB/FFFFFF?text=Pet+Shampoo&seed=1",
      "https://placehold.co/400x300/9370DB/FFFFFF?text=Pet+Shampoo&seed=2"
    ],
    dimensions: { height: 20, width: 8, length: 6, weight: 0.55 },
    brand: "PetClean Natural",
    colors: ["Transparente"],
    sku: "PCN-SHAMPOO-NEUTRAL-500ML",
  },
  {
    name: "Transportadora Grande com Rodinhas",
    description: "Transportadora resistente com rodinhas e alça telescópica. Ventilação 360°, abertura frontal e superior. Aprovada para viagens aéreas.",
    price: 599.90,
    rating: 4.7,
    category: "Transporte",
    disponibility: true,
    quantity_warehouse: 4,
    images: [
      "https://images.unsplash.com/photo-1598214960667-8c35c096dd3e?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1598214960667-8c35c096dd3e?w=400&h=300&fit=crop&q=80&seed=32",
      "https://images.unsplash.com/photo-1598214960667-8c35c096dd3e?w=400&h=300&fit=crop&q=80&seed=33"
    ],
    dimensions: { height: 35, width: 60, length: 40, weight: 6.5 },
    brand: "TravelPet Pro",
    colors: ["Preto", "Cinza"],
    sku: "TPP-CARRIER-WHEELS-L",
  }
];

// Insere os produtos realistas
db.getCollection(collectionName).insertMany(realisticProducts);

printLargeText("✅ Database with realistic pet products created successfully!", 8);
print(`📊 Inserted ${realisticProducts.length} realistic pet products`);
print("🖼️  All products include real images from Unsplash and fallback options");
print("🎯 Categories: Alimentação, Brinquedos, Acessórios, Higiene, Transporte, Gaiolas, Aquários");
print("🐕 Covers: Dogs, Cats, Birds, Fish and general pet products");
print("💰 Price range: R$ 24,90 to R$ 599,90");
print("⭐ All products have realistic ratings between 4.3 and 4.9");