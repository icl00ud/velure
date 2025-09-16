/**
 * Script para gerar produtos realistas de petshop
 * Inclui dados reais de produtos para pets com imagens de APIs gratuitas
 */

// Dados realistas de produtos de petshop
const petProducts = {
  dogs: {
    food: [
      {
        name: "Ração Premium Cães Adultos Frango e Arroz",
        description: "Ração super premium para cães adultos com frango real e arroz integral. Rico em proteínas e nutrientes essenciais para a saúde do seu pet.",
        price: 89.99,
        brand: "Royal Canin",
        category: "Alimentação",
        sku: "RC-DOG-ADULT-3KG"
      },
      {
        name: "Ração Natural Cães Filhotes Salmão",
        description: "Alimento natural para filhotes com salmão fresco, quinoa e vegetais. Sem conservantes artificiais, ideal para o desenvolvimento saudável.",
        price: 125.50,
        brand: "Farmina N&D",
        category: "Alimentação",
        sku: "FN-PUPPY-SAL-2.5KG"
      },
      {
        name: "Petisco Natural Osso de Couro",
        description: "Osso de couro natural 100% bovino, ideal para a higiene dental e entretenimento. Longa duração e sabor irresistível.",
        price: 15.99,
        brand: "DogChew",
        category: "Petiscos",
        sku: "DC-BONE-NATURAL"
      }
    ],
    toys: [
      {
        name: "Bola Interativa com Dispenser de Petiscos",
        description: "Bola de borracha resistente com compartimento para petiscos. Estimula o exercício mental e físico do seu cão.",
        price: 45.90,
        brand: "Kong",
        category: "Brinquedos",
        sku: "KONG-BALL-TREAT"
      },
      {
        name: "Corda Dental Tri-Nó Algodão",
        description: "Brinquedo de corda de algodão natural com três nós. Ajuda na limpeza dos dentes e fortalece a mandíbula.",
        price: 24.99,
        brand: "PetPlay",
        category: "Brinquedos",
        sku: "PP-ROPE-3KNOT"
      }
    ],
    accessories: [
      {
        name: "Coleira Ajustável Couro Legítimo",
        description: "Coleira de couro legítimo com fivela de metal resistente. Confortável e durável para uso diário.",
        price: 67.50,
        brand: "LeatherPet",
        category: "Acessórios",
        sku: "LP-COLLAR-LEATHER-M"
      },
      {
        name: "Cama Ortopédica Memory Foam",
        description: "Cama ortopédica com espuma memory foam para máximo conforto. Capa removível e lavável.",
        price: 189.90,
        brand: "ComfortPet",
        category: "Camas e Descanso",
        sku: "CP-BED-ORTHO-L"
      }
    ]
  },
  cats: {
    food: [
      {
        name: "Ração Premium Gatos Castrados Frango",
        description: "Ração especial para gatos castrados com controle de peso. Rica em proteínas e fibras, baixo teor de gordura.",
        price: 75.99,
        brand: "Hill's",
        category: "Alimentação",
        sku: "HILLS-CAT-CAST-1.5KG"
      },
      {
        name: "Sachê Gourmet Peixe ao Molho",
        description: "Alimento úmido gourmet com pedaços de peixe em molho saboroso. Rico em nutrientes e irresistível.",
        price: 3.50,
        brand: "Whiskas",
        category: "Alimentação",
        sku: "WK-SACHET-FISH"
      }
    ],
    toys: [
      {
        name: "Arranhador Torre com Brinquedos",
        description: "Torre arranhadora de sisal com plataformas e brinquedos suspensos. Ideal para exercício e diversão.",
        price: 159.90,
        brand: "CatTree",
        category: "Brinquedos",
        sku: "CT-TOWER-SCRATCH"
      },
      {
        name: "Varinha com Penas Naturais",
        description: "Brinquedo interativo com penas naturais coloridas. Estimula o instinto de caça dos felinos.",
        price: 19.99,
        brand: "FelinePlay",
        category: "Brinquedos",
        sku: "FP-WAND-FEATHER"
      }
    ],
    accessories: [
      {
        name: "Caixa de Areia Fechada com Filtro",
        description: "Caixa de areia fechada com sistema de filtro de odores. Fácil limpeza e máxima higiene.",
        price: 129.90,
        brand: "CleanCat",
        category: "Higiene",
        sku: "CC-LITTER-CLOSED"
      }
    ]
  },
  birds: [
    {
      name: "Mistura de Sementes Premium Canários",
      description: "Mistura especial de sementes selecionadas para canários. Rica em nutrientes e vitaminas essenciais.",
      price: 28.90,
      brand: "BirdSeed",
      category: "Alimentação",
      sku: "BS-CANARY-MIX-1KG"
    },
    {
      name: "Gaiola Spaciosa com Poleiros",
      description: "Gaiola espaçosa com poleiros de madeira natural e comedouros em inox. Design moderno e funcional.",
      price: 299.90,
      brand: "BirdHome",
      category: "Gaiolas",
      sku: "BH-CAGE-LARGE"
    }
  ],
  fish: [
    {
      name: "Ração em Flocos Peixes Tropicais",
      description: "Ração balanceada em flocos para peixes tropicais. Rica em proteínas e vitaminas para cores vibrantes.",
      price: 18.90,
      brand: "AquaFood",
      category: "Alimentação",
      sku: "AF-FLAKES-TROPICAL"
    },
    {
      name: "Filtro Submerso para Aquários",
      description: "Filtro submerso silencioso com bomba integrada. Ideal para aquários de 50 a 100 litros.",
      price: 89.50,
      brand: "AquaTech",
      category: "Aquários",
      sku: "AT-FILTER-SUB-100L"
    }
  ]
};

// URLs de imagens realistas usando APIs gratuitas
const imageApis = {
  unsplash: {
    dog: "https://images.unsplash.com/photo-1552053831-71594a27632d?w=400&h=300&fit=crop",
    cat: "https://images.unsplash.com/photo-1514888286974-6c03e2ca1dba?w=400&h=300&fit=crop",
    bird: "https://images.unsplash.com/photo-1452570053594-1b985d6ea890?w=400&h=300&fit=crop",
    fish: "https://images.unsplash.com/photo-1544551763-46a013bb70d5?w=400&h=300&fit=crop"
  },
  petImages: {
    dogFood: "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop",
    dogToy: "https://images.unsplash.com/photo-1601758228041-f3b2795255f1?w=400&h=300&fit=crop",
    dogAccessory: "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=400&h=300&fit=crop",
    catFood: "https://images.unsplash.com/photo-1571566882372-1598d88abd90?w=400&h=300&fit=crop",
    catToy: "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop",
    birdCage: "https://images.unsplash.com/photo-1555169062-013468b47731?w=400&h=300&fit=crop",
    aquarium: "https://images.unsplash.com/photo-1554263897-4bfa012dcac0?w=400&h=300&fit=crop"
  }
};

// Função para gerar dimensões realistas baseadas no tipo de produto
function generateDimensions(category, productName) {
  const dimensionsMap = {
    "Alimentação": { height: 25, width: 15, length: 10, weight: 2.0 },
    "Brinquedos": { height: 10, width: 10, length: 15, weight: 0.3 },
    "Acessórios": { height: 5, width: 20, length: 30, weight: 0.5 },
    "Camas e Descanso": { height: 15, width: 60, length: 80, weight: 2.5 },
    "Higiene": { height: 30, width: 40, length: 50, weight: 1.8 },
    "Gaiolas": { height: 50, width: 40, length: 30, weight: 5.0 },
    "Aquários": { height: 15, width: 30, length: 25, weight: 1.2 }
  };
  
  return dimensionsMap[category] || { height: 10, width: 10, length: 10, weight: 1.0 };
}

// Função para selecionar imagens baseadas no tipo de produto
function selectImages(category, productName) {
  const images = [];
  const baseUrl = imageApis.petImages;
  
  // Mapear categoria para tipo de imagem
  const categoryMap = {
    "Alimentação": productName.includes("Gato") || productName.includes("gato") ? baseUrl.catFood : baseUrl.dogFood,
    "Brinquedos": productName.includes("Gato") || productName.includes("gato") || productName.includes("Arranhador") ? baseUrl.catToy : baseUrl.dogToy,
    "Acessórios": baseUrl.dogAccessory,
    "Camas e Descanso": baseUrl.dogAccessory,
    "Higiene": baseUrl.catToy,
    "Gaiolas": baseUrl.birdCage,
    "Aquários": baseUrl.aquarium
  };
  
  const baseImage = categoryMap[category] || baseUrl.dogFood;
  
  // Gerar múltiplas variações da mesma imagem
  for (let i = 0; i < 3; i++) {
    images.push(`${baseImage}&seed=${Math.random()}`);
  }
  
  return images;
}

// Função para gerar cores baseadas no tipo de produto
function generateColors(category, productName) {
  const colorMap = {
    "Alimentação": ["Natural", "Marrom"],
    "Brinquedos": ["Vermelho", "Azul", "Verde", "Amarelo", "Rosa"],
    "Acessórios": ["Preto", "Marrom", "Azul", "Vermelho"],
    "Camas e Descanso": ["Cinza", "Bege", "Azul", "Marrom"],
    "Higiene": ["Branco", "Cinza", "Azul"],
    "Gaiolas": ["Branco", "Preto", "Prata"],
    "Aquários": ["Transparente", "Azul"]
  };
  
  const availableColors = colorMap[category] || ["Variadas"];
  const numColors = Math.min(Math.floor(Math.random() * 3) + 1, availableColors.length);
  
  return availableColors.slice(0, numColors);
}

// Função principal para gerar produtos
function generateRealisticProducts() {
  const products = [];
  let productId = 1;
  
  // Processar produtos de cães
  Object.keys(petProducts.dogs).forEach(subcategory => {
    petProducts.dogs[subcategory].forEach(product => {
      const generatedProduct = {
        name: product.name,
        description: product.description,
        price: product.price,
        rating: parseFloat((3.5 + Math.random() * 1.5).toFixed(1)), // Rating entre 3.5 e 5.0
        category: product.category,
        disponibility: Math.random() > 0.1, // 90% disponível
        quantity_warehouse: Math.floor(Math.random() * 100) + 10,
        images: selectImages(product.category, product.name),
        dimensions: generateDimensions(product.category, product.name),
        brand: product.brand,
        colors: generateColors(product.category, product.name),
        sku: product.sku,
      };
      products.push(generatedProduct);
    });
  });
  
  // Processar produtos de gatos
  Object.keys(petProducts.cats).forEach(subcategory => {
    petProducts.cats[subcategory].forEach(product => {
      const generatedProduct = {
        name: product.name,
        description: product.description,
        price: product.price,
        rating: parseFloat((3.5 + Math.random() * 1.5).toFixed(1)),
        category: product.category,
        disponibility: Math.random() > 0.1,
        quantity_warehouse: Math.floor(Math.random() * 100) + 10,
        images: selectImages(product.category, product.name),
        dimensions: generateDimensions(product.category, product.name),
        brand: product.brand,
        colors: generateColors(product.category, product.name),
        sku: product.sku,
      };
      products.push(generatedProduct);
    });
  });
  
  // Processar produtos de pássaros
  petProducts.birds.forEach(product => {
    const generatedProduct = {
      name: product.name,
      description: product.description,
      price: product.price,
      rating: parseFloat((3.5 + Math.random() * 1.5).toFixed(1)),
      category: product.category,
      disponibility: Math.random() > 0.1,
      quantity_warehouse: Math.floor(Math.random() * 100) + 10,
      images: selectImages(product.category, product.name),
      dimensions: generateDimensions(product.category, product.name),
      brand: product.brand,
      colors: generateColors(product.category, product.name),
      sku: product.sku,
    };
    products.push(generatedProduct);
  });
  
  // Processar produtos de peixes
  petProducts.fish.forEach(product => {
    const generatedProduct = {
      name: product.name,
      description: product.description,
      price: product.price,
      rating: parseFloat((3.5 + Math.random() * 1.5).toFixed(1)),
      category: product.category,
      disponibility: Math.random() > 0.1,
      quantity_warehouse: Math.floor(Math.random() * 100) + 10,
      images: selectImages(product.category, product.name),
      dimensions: generateDimensions(product.category, product.name),
      brand: product.brand,
      colors: generateColors(product.category, product.name),
      sku: product.sku,
    };
    products.push(generatedProduct);
  });
  
  return products;
}

// Exportar função para uso em outros scripts
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { generateRealisticProducts, petProducts, imageApis };
}

// Para teste direto do script
if (typeof window === 'undefined' && require.main === module) {
  const products = generateRealisticProducts();
  console.log(JSON.stringify(products, null, 2));
  console.log(`\nGerados ${products.length} produtos realistas!`);
}