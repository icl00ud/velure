/**
 * Script de inicialização do MongoDB com produtos realistas de petshop
 * Usa dados reais e imagens de APIs gratuitas
 */

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

  // === MAIS PRODUTOS PARA CÃES ===
  {
    name: "Guia Retrátil 5m com Trava de Segurança",
    description: "Guia retrátil automática com fita de nylon resistente. Sistema de trava com um toque, cabo ergonômico. Para cães até 25kg.",
    price: 79.90,
    rating: 4.5,
    category: "Acessórios",
    disponibility: true,
    quantity_warehouse: 41,
    images: [
      "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=400&h=300&fit=crop&q=80&seed=34",
      "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=400&h=300&fit=crop&q=80&seed=35",
      "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=400&h=300&fit=crop&q=80&seed=36"
    ],
    dimensions: { height: 12, width: 15, length: 4, weight: 0.35 },
    brand: "Flexi Pro",
    colors: ["Preto", "Azul", "Rosa"],
    sku: "FP-LEASH-RETRACT-5M",
  },
  {
    name: "Comedouro Lento Anti-Voracidade",
    description: "Comedouro com design labirinto que reduz a velocidade de alimentação. Previne problemas digestivos e obesidade. Antiderrapante.",
    price: 54.90,
    rating: 4.6,
    category: "Acessórios",
    disponibility: true,
    quantity_warehouse: 34,
    images: [
      "https://images.unsplash.com/photo-1623387641168-d9803ddd3f35?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1623387641168-d9803ddd3f35?w=400&h=300&fit=crop&q=80&seed=37",
      "https://images.unsplash.com/photo-1623387641168-d9803ddd3f35?w=400&h=300&fit=crop&q=80&seed=38"
    ],
    dimensions: { height: 6, width: 30, length: 30, weight: 0.6 },
    brand: "SlowFeeder",
    colors: ["Azul", "Verde", "Rosa"],
    sku: "SF-BOWL-SLOW-LARGE",
  },
  {
    name: "Ossinho Dental Filhotes 500g",
    description: "Petiscos mastigáveis para filhotes com cálcio e vitaminas. Formato especial para limpeza dental. Sabor frango.",
    price: 32.90,
    rating: 4.7,
    category: "Petiscos",
    disponibility: true,
    quantity_warehouse: 78,
    images: [
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop&q=80&seed=39",
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop&q=80&seed=40",
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop&q=80&seed=41"
    ],
    dimensions: { height: 20, width: 15, length: 8, weight: 0.5 },
    brand: "DentaPet",
    colors: ["Natural"],
    sku: "DP-DENTAL-PUPPY-500G",
  },
  {
    name: "Tapete Higiênico Super Absorvente 30un",
    description: "Tapetes higiênicos com gel super absorvente e neutralizador de odores. Bordas adesivas para fixação. Ideal para treinamento.",
    price: 45.90,
    rating: 4.4,
    category: "Higiene",
    disponibility: true,
    quantity_warehouse: 52,
    images: [
      "https://placehold.co/400x300/87CEEB/000000?text=Puppy+Pads",
      "https://placehold.co/400x300/87CEEB/000000?text=Puppy+Pads&seed=1",
      "https://placehold.co/400x300/87CEEB/000000?text=Puppy+Pads&seed=2"
    ],
    dimensions: { height: 8, width: 40, length: 30, weight: 1.8 },
    brand: "CleanPad Pro",
    colors: ["Branco"],
    sku: "CPP-PAD-SUPER-30UN",
  },
  {
    name: "Roupinha Fleece para Cães Pequenos",
    description: "Roupinha de fleece macia e quentinha. Fácil de vestir com velcro, disponível em diversos tamanhos. Lavável na máquina.",
    price: 49.90,
    rating: 4.3,
    category: "Vestuário",
    disponibility: true,
    quantity_warehouse: 27,
    images: [
      "https://images.unsplash.com/photo-1548199973-03cce0bbc87b?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1548199973-03cce0bbc87b?w=400&h=300&fit=crop&q=80&seed=42",
      "https://images.unsplash.com/photo-1548199973-03cce0bbc87b?w=400&h=300&fit=crop&q=80&seed=43"
    ],
    dimensions: { height: 2, width: 30, length: 25, weight: 0.15 },
    brand: "PetFashion",
    colors: ["Rosa", "Azul", "Vermelho", "Cinza"],
    sku: "PF-FLEECE-SMALL",
  },

  // === MAIS PRODUTOS PARA GATOS ===
  {
    name: "Fonte de Água Automática com Filtro",
    description: "Fonte de água circulante com sistema de filtração triplo. Estimula os gatos a beberem mais água. Capacidade 2 litros, ultra silenciosa.",
    price: 159.90,
    rating: 4.8,
    category: "Acessórios",
    disponibility: true,
    quantity_warehouse: 19,
    images: [
      "https://images.unsplash.com/photo-1573865526739-10c1dd85fd5f?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1573865526739-10c1dd85fd5f?w=400&h=300&fit=crop&q=80&seed=44",
      "https://images.unsplash.com/photo-1573865526739-10c1dd85fd5f?w=400&h=300&fit=crop&q=80&seed=45"
    ],
    dimensions: { height: 15, width: 22, length: 22, weight: 1.2 },
    brand: "CatFlow",
    colors: ["Branco", "Cinza"],
    sku: "CF-FOUNTAIN-2L",
  },
  {
    name: "Areia Sanitária Aglomerante Perfumada 4kg",
    description: "Areia sanitária super aglomerante com fragrância de lavanda. Alto poder de absorção, controle superior de odores. Baixo pó.",
    price: 39.90,
    rating: 4.5,
    category: "Higiene",
    disponibility: true,
    quantity_warehouse: 65,
    images: [
      "https://placehold.co/400x300/DDA0DD/000000?text=Cat+Litter",
      "https://placehold.co/400x300/DDA0DD/000000?text=Cat+Litter&seed=1",
      "https://placehold.co/400x300/DDA0DD/000000?text=Cat+Litter&seed=2"
    ],
    dimensions: { height: 30, width: 25, length: 10, weight: 4.0 },
    brand: "LitterFresh",
    colors: ["Lavanda"],
    sku: "LF-LITTER-CLUMP-4KG",
  },
  {
    name: "Túnel Dobrável para Gatos 3 vias",
    description: "Túnel de brincar dobrável com 3 entradas e bola suspensa. Material durável, fácil armazenamento. Ideal para exercício e diversão.",
    price: 89.90,
    rating: 4.6,
    category: "Brinquedos",
    disponibility: true,
    quantity_warehouse: 23,
    images: [
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop&q=80&seed=46",
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop&q=80&seed=47",
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop&q=80&seed=48"
    ],
    dimensions: { height: 30, width: 120, length: 30, weight: 0.8 },
    brand: "PlayCat",
    colors: ["Cinza", "Bege"],
    sku: "PC-TUNNEL-3WAY",
  },
  {
    name: "Escova Removedora de Pelos",
    description: "Escova profissional para remoção de pelos mortos. Lâminas de aço inoxidável, cabo ergonômico. Reduz em 90% a queda de pelos.",
    price: 67.90,
    rating: 4.7,
    category: "Higiene",
    disponibility: true,
    quantity_warehouse: 38,
    images: [
      "https://placehold.co/400x300/FF69B4/FFFFFF?text=Pet+Brush",
      "https://placehold.co/400x300/FF69B4/FFFFFF?text=Pet+Brush&seed=1",
      "https://placehold.co/400x300/FF69B4/FFFFFF?text=Pet+Brush&seed=2"
    ],
    dimensions: { height: 18, width: 10, length: 5, weight: 0.25 },
    brand: "FurRemover Pro",
    colors: ["Azul", "Rosa"],
    sku: "FRP-BRUSH-DESHED",
  },
  {
    name: "Rede Suspensa para Janela",
    description: "Rede de descanso para gatos com ventosas ultra forte. Suporta até 15kg, tecido respirável. Perfeita para banho de sol.",
    price: 79.90,
    rating: 4.4,
    category: "Camas e Descanso",
    disponibility: true,
    quantity_warehouse: 16,
    images: [
      "https://images.unsplash.com/photo-1558617047-ac1a6b5abbd7?w=400&h=300&fit=crop&q=80&seed=49",
      "https://images.unsplash.com/photo-1558617047-ac1a6b5abbd7?w=400&h=300&fit=crop&q=80&seed=50",
      "https://images.unsplash.com/photo-1558617047-ac1a6b5abbd7?w=400&h=300&fit=crop&q=80&seed=51"
    ],
    dimensions: { height: 5, width: 50, length: 40, weight: 0.6 },
    brand: "WindowCat",
    colors: ["Cinza", "Bege"],
    sku: "WC-HAMMOCK-WINDOW",
  },

  // === PRODUTOS PARA PEQUENOS ROEDORES ===
  {
    name: "Gaiola Ampla Hamster 2 Andares",
    description: "Habitat completo para hamster com 2 andares conectados. Inclui roda de exercício, bebedouro e comedouro. Base alta anti-derramamento.",
    price: 189.90,
    rating: 4.5,
    category: "Gaiolas",
    disponibility: true,
    quantity_warehouse: 11,
    images: [
      "https://placehold.co/400x300/FFD700/000000?text=Hamster+Cage",
      "https://placehold.co/400x300/FFD700/000000?text=Hamster+Cage&seed=1",
      "https://placehold.co/400x300/FFD700/000000?text=Hamster+Cage&seed=2"
    ],
    dimensions: { height: 40, width: 60, length: 35, weight: 5.5 },
    brand: "SmallPet Home",
    colors: ["Transparente"],
    sku: "SPH-CAGE-HAMSTER-2FL",
  },
  {
    name: "Ração Premium para Coelhos 1.5kg",
    description: "Ração extrusada para coelhos adultos. Rica em fibras, com feno timothy e vegetais. Promove desgaste dental adequado.",
    price: 42.90,
    rating: 4.6,
    category: "Alimentação",
    disponibility: true,
    quantity_warehouse: 47,
    images: [
      "https://images.unsplash.com/photo-1585110396000-c9ffd4e4b308?w=400&h=300&fit=crop&q=80",
      "https://images.unsplash.com/photo-1585110396000-c9ffd4e4b308?w=400&h=300&fit=crop&q=80&seed=52",
      "https://images.unsplash.com/photo-1585110396000-c9ffd4e4b308?w=400&h=300&fit=crop&q=80&seed=53"
    ],
    dimensions: { height: 25, width: 18, length: 8, weight: 1.5 },
    brand: "BunnyFood Premium",
    colors: ["Natural"],
    sku: "BFP-RABBIT-ADULT-1.5KG",
  },
  {
    name: "Roda de Exercício Silenciosa 20cm",
    description: "Roda de exercício silenciosa para hamsters e pequenos roedores. Base sólida, sem eixo central. Material atóxico e seguro.",
    price: 54.90,
    rating: 4.7,
    category: "Brinquedos",
    disponibility: true,
    quantity_warehouse: 31,
    images: [
      "https://placehold.co/400x300/FF6347/FFFFFF?text=Exercise+Wheel",
      "https://placehold.co/400x300/FF6347/FFFFFF?text=Exercise+Wheel&seed=1",
      "https://placehold.co/400x300/FF6347/FFFFFF?text=Exercise+Wheel&seed=2"
    ],
    dimensions: { height: 20, width: 20, length: 8, weight: 0.3 },
    brand: "RunWheel Silent",
    colors: ["Rosa", "Azul", "Verde"],
    sku: "RWS-WHEEL-20CM",
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
  },
  {
    name: "Manta Térmica Pet com Desligamento Automático",
    description: "Manta térmica elétrica com controle de temperatura e timer. Capa impermeável removível e lavável. Segurança certificada.",
    price: 129.90,
    rating: 4.6,
    category: "Camas e Descanso",
    disponibility: true,
    quantity_warehouse: 14,
    images: [
      "https://placehold.co/400x300/FF8C00/FFFFFF?text=Heated+Blanket",
      "https://placehold.co/400x300/FF8C00/FFFFFF?text=Heated+Blanket&seed=1",
      "https://placehold.co/400x300/FF8C00/FFFFFF?text=Heated+Blanket&seed=2"
    ],
    dimensions: { height: 5, width: 60, length: 45, weight: 1.2 },
    brand: "WarmPet",
    colors: ["Cinza", "Marrom"],
    sku: "WP-BLANKET-HEATED-M",
  },
  {
    name: "Kit Primeiros Socorros Pet",
    description: "Kit completo de primeiros socorros para emergências. Inclui bandagens, gazes, termômetro, luvas e manual de instruções.",
    price: 89.90,
    rating: 4.8,
    category: "Saúde",
    disponibility: true,
    quantity_warehouse: 25,
    images: [
      "https://placehold.co/400x300/DC143C/FFFFFF?text=First+Aid+Kit",
      "https://placehold.co/400x300/DC143C/FFFFFF?text=First+Aid+Kit&seed=1",
      "https://placehold.co/400x300/DC143C/FFFFFF?text=First+Aid+Kit&seed=2"
    ],
    dimensions: { height: 10, width: 25, length: 18, weight: 0.8 },
    brand: "PetCare Emergency",
    colors: ["Vermelho"],
    sku: "PCE-FIRSTAID-KIT",
  },
  {
    name: "Localizador GPS para Coleira",
    description: "Rastreador GPS em tempo real com app móvel. Bateria de longa duração, resistente à água. Histórico de localização e zona segura.",
    price: 249.90,
    rating: 4.5,
    category: "Tecnologia",
    disponibility: true,
    quantity_warehouse: 18,
    images: [
      "https://placehold.co/400x300/4169E1/FFFFFF?text=GPS+Tracker",
      "https://placehold.co/400x300/4169E1/FFFFFF?text=GPS+Tracker&seed=1",
      "https://placehold.co/400x300/4169E1/FFFFFF?text=GPS+Tracker&seed=2"
    ],
    dimensions: { height: 2, width: 5, length: 4, weight: 0.05 },
    brand: "TrackPet Pro",
    colors: ["Preto"],
    sku: "TPP-GPS-TRACKER",
  }
];

// Insere os produtos realistas
db.getCollection(collectionName).insertMany(realisticProducts);

printLargeText("✅ Database with realistic pet products created successfully!", 8);
print(`📊 Inserted ${realisticProducts.length} realistic pet products`);
print("🖼️  All products include real images from Unsplash and fallback options");
print("🎯 Categories: Alimentação, Brinquedos, Acessórios, Higiene, Transporte, Gaiolas, Aquários, Petiscos, Vestuário, Suplementos, Saúde, Tecnologia, Camas e Descanso");
print("🐕 Covers: Dogs, Cats, Birds, Fish, Small Rodents and general pet products");
print("💰 Price range: R$ 24,90 to R$ 599,90");
print("⭐ All products have realistic ratings between 4.3 and 4.9");