// Script para popular produtos no MongoDB
// Execução: docker exec -i mongodb mongosh -u velure_user -p velure_password --authenticationDatabase admin < populate-products.js

// Conectar ao banco correto
db = db.getSiblingDB('product_service');

// Limpar produtos existentes (opcional)
db.products.deleteMany({});

// Produtos realistas para pet shop
const realisticProducts = [
  {
    name: "Ração Premium para Cães Adultos",
    description: "Ração completa e balanceada com proteínas de alta qualidade, ideal para cães adultos de todas as raças. Contém vitaminas e minerais essenciais.",
    price: 189.90,
    rating: 4.7,
    category: "Alimentação",
    quantity: 150,
    images: [
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=500",
      "https://images.unsplash.com/photo-1628009368231-7bb7cfcb0def?w=500"
    ],
    dimensions: {
      weight: 15,
      height: 60,
      width: 40,
      length: 15
    },
    brand: "PetNutri",
    colors: [],
    sku: "RCP-ADT-15KG",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Arranhador para Gatos com Brinquedos",
    description: "Arranhador vertical com sisal resistente, plataformas de descanso e brinquedos suspensos. Perfeito para gatos brincalhões.",
    price: 299.90,
    rating: 4.8,
    category: "Brinquedos",
    quantity: 45,
    images: [
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=500",
      "https://images.unsplash.com/photo-1574144611937-0df059b5ef3e?w=500"
    ],
    dimensions: {
      weight: 8,
      height: 120,
      width: 40,
      length: 40
    },
    brand: "FelineFun",
    colors: ["Bege", "Cinza"],
    sku: "ARR-GAT-120",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Alpiste Premium para Pássaros",
    description: "Mistura especial de sementes selecionadas para canários e outros pássaros pequenos. Rico em nutrientes.",
    price: 34.90,
    rating: 4.5,
    category: "Alimentação",
    quantity: 200,
    images: [
      "https://images.unsplash.com/photo-1552728089-57bdde30beb3?w=500"
    ],
    dimensions: {
      weight: 0.5,
      height: 25,
      width: 15,
      length: 8
    },
    brand: "BirdLife",
    colors: [],
    sku: "ALP-PAS-500G",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Aquário Completo 100 Litros",
    description: "Kit aquário completo com filtro, bomba de ar, iluminação LED e termômetro. Ideal para peixes tropicais.",
    price: 589.90,
    rating: 4.9,
    category: "Acessórios",
    quantity: 25,
    images: [
      "https://images.unsplash.com/photo-1520366498724-709889c0c685?w=500",
      "https://images.unsplash.com/photo-1524704796725-9fc3044a58b1?w=500"
    ],
    dimensions: {
      weight: 15,
      height: 50,
      width: 80,
      length: 35
    },
    brand: "AquaPro",
    colors: ["Preto"],
    sku: "AQU-100L-KIT",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Coleira Antipulgas para Cães",
    description: "Coleira com ação prolongada de até 8 meses, repele pulgas, carrapatos e mosquitos. À prova d'água.",
    price: 79.90,
    rating: 4.6,
    category: "Saúde",
    quantity: 180,
    images: [
      "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=500"
    ],
    dimensions: {
      weight: 0.05,
      height: 2,
      width: 60,
      length: 2
    },
    brand: "PetProtect",
    colors: ["Cinza"],
    sku: "COL-APG-8M",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Cama Ortopédica para Cães Grandes",
    description: "Cama com espuma de memória ortopédica, capa removível lavável, ideal para cães idosos ou com problemas articulares.",
    price: 349.90,
    rating: 4.8,
    category: "Conforto",
    quantity: 35,
    images: [
      "https://images.unsplash.com/photo-1581888227599-779811939961?w=500",
      "https://images.unsplash.com/photo-1615751072497-5f5169febe17?w=500"
    ],
    dimensions: {
      weight: 3,
      height: 15,
      width: 100,
      length: 80
    },
    brand: "ComfortPet",
    colors: ["Cinza", "Marrom", "Azul"],
    sku: "CAM-ORT-GG",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Bebedouro Automático 6 Litros",
    description: "Fonte de água com filtro triplo, circulação contínua, ultra silenciosa. Estimula pets a beberem mais água.",
    price: 159.90,
    rating: 4.7,
    category: "Acessórios",
    quantity: 90,
    images: [
      "https://images.unsplash.com/photo-1591696205602-2f950c417cb9?w=500"
    ],
    dimensions: {
      weight: 1.2,
      height: 18,
      width: 22,
      length: 22
    },
    brand: "HydroPet",
    colors: ["Branco", "Azul"],
    sku: "BEB-AUT-6L",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Gaiola Luxo para Hamster",
    description: "Gaiola espaçosa com 3 andares, escadas, roda de exercícios e comedouros inclusos. Fácil limpeza.",
    price: 249.90,
    rating: 4.5,
    category: "Habitação",
    quantity: 40,
    images: [
      "https://images.unsplash.com/photo-1425082661705-1834bfd09dca?w=500"
    ],
    dimensions: {
      weight: 4,
      height: 45,
      width: 60,
      length: 40
    },
    brand: "RodentHome",
    colors: ["Azul", "Rosa", "Verde"],
    sku: "GAI-HAM-3A",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Petiscos Naturais para Gatos",
    description: "Snacks de frango desidratado 100% natural, sem conservantes, corantes ou aromatizantes artificiais.",
    price: 24.90,
    rating: 4.9,
    category: "Alimentação",
    quantity: 300,
    images: [
      "https://images.unsplash.com/photo-1589883661923-6476cb0ae9f2?w=500"
    ],
    dimensions: {
      weight: 0.08,
      height: 15,
      width: 10,
      length: 3
    },
    brand: "NaturalPet",
    colors: [],
    sku: "PET-GAT-FRA",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Shampoo Hipoalergênico para Cães",
    description: "Fórmula suave desenvolvida para peles sensíveis, sem parabenos, com extratos naturais de camomila e aloe vera.",
    price: 45.90,
    rating: 4.6,
    category: "Higiene",
    quantity: 120,
    images: [
      "https://images.unsplash.com/photo-1585559700398-1385b3a8aeb6?w=500"
    ],
    dimensions: {
      weight: 0.5,
      height: 20,
      width: 6,
      length: 6
    },
    brand: "PetClean",
    colors: [],
    sku: "SHA-HIP-500ML",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Bolinha de Tênis para Cães",
    description: "Pacote com 3 bolas de tênis resistentes, perfeitas para brincadeiras de buscar. Material atóxico.",
    price: 29.90,
    rating: 4.7,
    category: "Brinquedos",
    quantity: 250,
    images: [
      "https://images.unsplash.com/photo-1587300003388-59208cc962cb?w=500"
    ],
    dimensions: {
      weight: 0.15,
      height: 6.5,
      width: 6.5,
      length: 6.5
    },
    brand: "PlayDog",
    colors: ["Amarelo/Verde"],
    sku: "BOL-TEN-3UN",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Casinha de Madeira para Cães Médios",
    description: "Casa de madeira tratada com telhado impermeável, pés elevados para isolamento térmico, fácil montagem.",
    price: 449.90,
    rating: 4.8,
    category: "Habitação",
    quantity: 20,
    images: [
      "https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=500"
    ],
    dimensions: {
      weight: 18,
      height: 85,
      width: 75,
      length: 90
    },
    brand: "WoodPet",
    colors: ["Natural"],
    sku: "CAS-MAD-M",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Areia Sanitária Aglomerante",
    description: "Areia super absorvente que forma blocos sólidos facilitando a limpeza. Controla odores por até 30 dias.",
    price: 54.90,
    rating: 4.5,
    category: "Higiene",
    quantity: 180,
    images: [
      "https://images.unsplash.com/photo-1615751072497-5f5169febe17?w=500"
    ],
    dimensions: {
      weight: 4,
      height: 10,
      width: 30,
      length: 40
    },
    brand: "CleanCat",
    colors: [],
    sku: "ARE-AGL-4KG",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Ração para Peixes Tropicais",
    description: "Alimento em flocos balanceado com vitaminas e minerais essenciais. Realça as cores naturais dos peixes.",
    price: 28.90,
    rating: 4.6,
    category: "Alimentação",
    quantity: 160,
    images: [
      "https://images.unsplash.com/photo-1535591273668-578e31182c4f?w=500"
    ],
    dimensions: {
      weight: 0.1,
      height: 12,
      width: 8,
      length: 4
    },
    brand: "AquaFood",
    colors: [],
    sku: "RAC-PEI-TRP",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Transportadora para Gatos",
    description: "Caixa de transporte resistente com porta de metal, ventilação 360°, alça confortável e trava de segurança.",
    price: 129.90,
    rating: 4.7,
    category: "Transporte",
    quantity: 65,
    images: [
      "https://images.unsplash.com/photo-1589883661923-6476cb0ae9f2?w=500"
    ],
    dimensions: {
      weight: 1.5,
      height: 32,
      width: 48,
      length: 32
    },
    brand: "SafeTravel",
    colors: ["Azul", "Rosa", "Cinza"],
    sku: "TRA-GAT-M",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Comedouro Antivoracidade",
    description: "Tigela com design especial que reduz a velocidade de ingestão, previne engasgos e melhora a digestão.",
    price: 69.90,
    rating: 4.8,
    category: "Acessórios",
    quantity: 140,
    images: [
      "https://images.unsplash.com/photo-1591696205602-2f950c417cb9?w=500"
    ],
    dimensions: {
      weight: 0.4,
      height: 8,
      width: 25,
      length: 25
    },
    brand: "SlowEat",
    colors: ["Cinza", "Verde", "Azul"],
    sku: "COM-ANT-M",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Brinquedo Interativo para Cães",
    description: "Dispensador de petiscos com níveis de dificuldade ajustáveis. Estimula a inteligência e mantém o pet ocupado.",
    price: 89.90,
    rating: 4.7,
    category: "Brinquedos",
    quantity: 95,
    images: [
      "https://images.unsplash.com/photo-1587300003388-59208cc962cb?w=500"
    ],
    dimensions: {
      weight: 0.35,
      height: 15,
      width: 15,
      length: 15
    },
    brand: "SmartPet",
    colors: ["Laranja", "Roxo"],
    sku: "BRI-INT-M",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Escova Removedora de Pelos",
    description: "Escova com cerdas de aço inoxidável, remove pelos mortos sem machucar, ideal para todas as raças.",
    price: 49.90,
    rating: 4.9,
    category: "Higiene",
    quantity: 200,
    images: [
      "https://images.unsplash.com/photo-1585559700398-1385b3a8aeb6?w=500"
    ],
    dimensions: {
      weight: 0.15,
      height: 20,
      width: 10,
      length: 5
    },
    brand: "GroomPro",
    colors: ["Azul", "Rosa"],
    sku: "ESC-PEL-UNI",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Osso de Nylon para Mordida",
    description: "Osso resistente para cães de médio a grande porte, ajuda na limpeza dos dentes e fortalece a mandíbula.",
    price: 39.90,
    rating: 4.6,
    category: "Brinquedos",
    quantity: 170,
    images: [
      "https://images.unsplash.com/photo-1589883661923-6476cb0ae9f2?w=500"
    ],
    dimensions: {
      weight: 0.2,
      height: 15,
      width: 5,
      length: 5
    },
    brand: "ChewMaster",
    colors: ["Branco"],
    sku: "OSO-NYL-M",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Vitaminas para Cães Idosos",
    description: "Suplemento com condroitina, glucosamina e ômega 3. Suporta articulações saudáveis e pelagem brilhante.",
    price: 119.90,
    rating: 4.8,
    category: "Saúde",
    quantity: 85,
    images: [
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=500"
    ],
    dimensions: {
      weight: 0.15,
      height: 12,
      width: 6,
      length: 6
    },
    brand: "VitalPet",
    colors: [],
    sku: "VIT-CAO-IDO",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Guia Retrátil 5 Metros",
    description: "Guia com trava de segurança, cabo reforçado, suporta até 50kg, alça ergonômica com sistema anti-derrapante.",
    price: 99.90,
    rating: 4.7,
    category: "Acessórios",
    quantity: 110,
    images: [
      "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=500"
    ],
    dimensions: {
      weight: 0.3,
      height: 12,
      width: 15,
      length: 5
    },
    brand: "WalkSafe",
    colors: ["Preto", "Azul", "Vermelho"],
    sku: "GUI-RET-5M",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Ração Úmida Premium para Gatos",
    description: "Sachê com pedaços de carne ao molho, sem grãos, alta palatabilidade, rica em proteínas. Pack com 12 unidades.",
    price: 79.90,
    rating: 4.9,
    category: "Alimentação",
    quantity: 130,
    images: [
      "https://images.unsplash.com/photo-1589883661923-6476cb0ae9f2?w=500"
    ],
    dimensions: {
      weight: 1,
      height: 12,
      width: 18,
      length: 24
    },
    brand: "FelineGourmet",
    colors: [],
    sku: "RAC-UMI-12UN",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Tapete Higiênico Super Absorvente",
    description: "Tapetes descartáveis com tecnologia de gel neutralizador de odores. Pack com 30 unidades 60x60cm.",
    price: 89.90,
    rating: 4.5,
    category: "Higiene",
    quantity: 150,
    images: [
      "https://images.unsplash.com/photo-1615751072497-5f5169febe17?w=500"
    ],
    dimensions: {
      weight: 2,
      height: 5,
      width: 30,
      length: 40
    },
    brand: "CleanPad",
    colors: [],
    sku: "TAP-HIG-30UN",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Poleiro Natural para Pássaros",
    description: "Galho natural com textura irregular, estimula o desgaste natural das unhas, fácil instalação em gaiolas.",
    price: 19.90,
    rating: 4.6,
    category: "Acessórios",
    quantity: 220,
    images: [
      "https://images.unsplash.com/photo-1552728089-57bdde30beb3?w=500"
    ],
    dimensions: {
      weight: 0.05,
      height: 25,
      width: 2,
      length: 2
    },
    brand: "BirdNature",
    colors: ["Natural"],
    sku: "POL-NAT-25CM",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Roupinha de Inverno para Cães Pequenos",
    description: "Casaco forrado em fleece, impermeável, com abertura para guia, disponível em diversos tamanhos.",
    price: 69.90,
    rating: 4.7,
    category: "Conforto",
    quantity: 100,
    images: [
      "https://images.unsplash.com/photo-1581888227599-779811939961?w=500"
    ],
    dimensions: {
      weight: 0.12,
      height: 1,
      width: 25,
      length: 30
    },
    brand: "PetFashion",
    colors: ["Vermelho", "Azul", "Rosa", "Cinza"],
    sku: "ROU-INV-PP",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Filtro Externo para Aquário",
    description: "Filtro biológico com 3 estágios de filtragem, vazão de 1000L/h, silencioso, ideal para aquários de 100 a 200L.",
    price: 389.90,
    rating: 4.8,
    category: "Acessórios",
    quantity: 30,
    images: [
      "https://images.unsplash.com/photo-1520366498724-709889c0c685?w=500"
    ],
    dimensions: {
      weight: 2.5,
      height: 35,
      width: 20,
      length: 15
    },
    brand: "AquaFilter",
    colors: ["Preto"],
    sku: "FIL-EXT-1000",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Vermífugo de Amplo Espectro",
    description: "Comprimidos palatáveis para cães e gatos, elimina vermes intestinais, dose única conforme peso do animal.",
    price: 45.90,
    rating: 4.9,
    category: "Saúde",
    quantity: 190,
    images: [
      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=500"
    ],
    dimensions: {
      weight: 0.02,
      height: 8,
      width: 10,
      length: 2
    },
    brand: "PetHealth",
    colors: [],
    sku: "VER-AMP-4CP",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Túnel para Gatos Dobrável",
    description: "Túnel de brinquedo com 3 vias e bolas suspensas, dobrável para fácil armazenamento, tecido resistente.",
    price: 99.90,
    rating: 4.6,
    category: "Brinquedos",
    quantity: 75,
    images: [
      "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=500"
    ],
    dimensions: {
      weight: 0.5,
      height: 25,
      width: 25,
      length: 120
    },
    brand: "PlayCat",
    colors: ["Cinza", "Azul"],
    sku: "TUN-GAT-3V",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Comedouro Automático Programável",
    description: "Dispensa ração automaticamente em até 4 horários por dia, capacidade para 6L, funciona com pilhas ou energia.",
    price: 299.90,
    rating: 4.8,
    category: "Acessórios",
    quantity: 50,
    images: [
      "https://images.unsplash.com/photo-1591696205602-2f950c417cb9?w=500"
    ],
    dimensions: {
      weight: 2,
      height: 35,
      width: 30,
      length: 30
    },
    brand: "AutoFeed",
    colors: ["Branco"],
    sku: "COM-AUT-6L",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    name: "Feno Premium para Coelhos",
    description: "Feno de capim timothy de primeira corte, rico em fibras, essencial para a saúde digestiva. Pacote 500g.",
    price: 39.90,
    rating: 4.7,
    category: "Alimentação",
    quantity: 140,
    images: [
      "https://images.unsplash.com/photo-1585110396000-c9ffd4e4b308?w=500"
    ],
    dimensions: {
      weight: 0.5,
      height: 30,
      width: 25,
      length: 10
    },
    brand: "BunnyFood",
    colors: [],
    sku: "FEN-COE-500G",
    createdAt: new Date(),
    updatedAt: new Date()
  }
];

// Inserir produtos
const result = db.products.insertMany(realisticProducts);

print("✅ Produtos inseridos com sucesso!");
print(`📦 Total de produtos: ${result.insertedIds.length}`);
print(`🏪 Banco de dados: ${db.getName()}`);
print(`📊 Coleção: products`);

// Verificar inserção
const count = db.products.countDocuments();
print(`\n✓ Verificação: ${count} produtos na coleção`);
