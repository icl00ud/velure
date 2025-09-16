/**
 * Serviço para buscar imagens reais de produtos de pets
 * Utiliza múltiplas APIs gratuitas com fallback
 */

// Configurações das APIs de imagem
const IMAGE_APIS = {
  unsplash: {
    baseUrl: 'https://images.unsplash.com',
    rateLimited: false,
    categories: {
      dogs: [
        'photo-1552053831-71594a27632d', // Golden Retriever
        'photo-1583337130417-3346a1be7dee', // Dog with collar
        'photo-1601758228041-f3b2795255f1', // Dog toys
        'photo-1589924691995-400dc9ecc119', // Dog food
        'photo-1548199973-03cce0bbc87b', // Dog portrait
        'photo-1587300003388-59208cc962cb', // Cute puppy
      ],
      cats: [
        'photo-1514888286974-6c03e2ca1dba', // Orange cat
        'photo-1571566882372-1598d88abd90', // Cat food
        'photo-1545249390-6bdfa286032f', // Cat with toy
        'photo-1596854407944-bf87f6fdd49e', // Cat portrait
        'photo-1533738363-b7f9aef128ce', // Cat playing
      ],
      birds: [
        'photo-1452570053594-1b985d6ea890', // Colorful bird
        'photo-1555169062-013468b47731', // Bird cage
        'photo-1598300042247-d088f8ab3a91', // Bird seed
        'photo-1594736797933-d0601ba2fe65', // Parrot
      ],
      fish: [
        'photo-1544551763-46a013bb70d5', // Aquarium
        'photo-1554263897-4bfa012dcac0', // Fish tank
        'photo-1520637836862-4d197d17c55a', // Tropical fish
        'photo-1559827260-dc66d52bef19', // Fish food
      ],
      petAccessories: [
        'photo-1583337130417-3346a1be7dee', // Dog collar
        'photo-1558617047-ac1a6b5abbd7', // Pet bed
        'photo-1574158622682-e40e69881006', // Pet leash
        'photo-1598214960667-8c35c096dd3e', // Pet carrier
      ]
    }
  },
  picsum: {
    baseUrl: 'https://picsum.photos',
    rateLimited: false,
    seeds: {
      pets: [1, 23, 45, 67, 89, 123, 456, 789, 101, 202]
    }
  },
  placeholder: {
    baseUrl: 'https://placehold.co',
    rateLimited: false,
    colors: ['8B4513', '228B22', '4169E1', 'FF6347', '9370DB']
  }
};

// Cache para URLs de imagens geradas
const imageCache = new Map();

class PetImageService {
  constructor() {
    this.apiRotation = 0;
    this.requestCount = 0;
  }

  /**
   * Gera URLs de imagens para um produto específico
   * @param {string} category - Categoria do produto (Alimentação, Brinquedos, etc.)
   * @param {string} productName - Nome do produto
   * @param {string} animalType - Tipo de animal (dogs, cats, birds, fish)
   * @param {number} count - Número de imagens desejadas (padrão: 3)
   * @returns {Array} Array de URLs de imagens
   */
  generateProductImages(category, productName, animalType = 'dogs', count = 3) {
    const cacheKey = `${category}-${animalType}-${productName}`;
    
    if (imageCache.has(cacheKey)) {
      return imageCache.get(cacheKey);
    }

    const images = [];
    
    for (let i = 0; i < count; i++) {
      const imageUrl = this.selectImageStrategy(category, productName, animalType, i);
      images.push(imageUrl);
    }
    
    imageCache.set(cacheKey, images);
    return images;
  }

  /**
   * Seleciona a estratégia de imagem baseada no tipo de produto
   */
  selectImageStrategy(category, productName, animalType, index) {
    // Estratégia baseada na categoria
    if (this.isFood(category, productName)) {
      return this.getFoodImage(animalType, index);
    } else if (this.isToy(category, productName)) {
      return this.getToyImage(animalType, index);
    } else if (this.isAccessory(category, productName)) {
      return this.getAccessoryImage(animalType, index);
    } else if (this.isHousing(category, productName)) {
      return this.getHousingImage(animalType, index);
    }
    
    // Fallback para imagem genérica do animal
    return this.getGenericAnimalImage(animalType, index);
  }

  /**
   * Verifica se é produto de alimentação
   */
  isFood(category, productName) {
    return category === 'Alimentação' || 
           productName.toLowerCase().includes('ração') ||
           productName.toLowerCase().includes('petisco') ||
           productName.toLowerCase().includes('sachê');
  }

  /**
   * Verifica se é brinquedo
   */
  isToy(category, productName) {
    return category === 'Brinquedos' ||
           productName.toLowerCase().includes('bola') ||
           productName.toLowerCase().includes('corda') ||
           productName.toLowerCase().includes('brinquedo');
  }

  /**
   * Verifica se é acessório
   */
  isAccessory(category, productName) {
    return category === 'Acessórios' ||
           productName.toLowerCase().includes('coleira') ||
           productName.toLowerCase().includes('cama') ||
           productName.toLowerCase().includes('tigela');
  }

  /**
   * Verifica se é habitação (gaiolas, aquários, etc.)
   */
  isHousing(category, productName) {
    return category === 'Gaiolas' ||
           category === 'Aquários' ||
           productName.toLowerCase().includes('gaiola') ||
           productName.toLowerCase().includes('aquário');
  }

  /**
   * Gera URL para imagem de comida
   */
  getFoodImage(animalType, index) {
    const unsplashIds = IMAGE_APIS.unsplash.categories[animalType];
    if (unsplashIds && unsplashIds.length > 0) {
      const foodIds = unsplashIds.filter(id => 
        id.includes('food') || id.includes('589924691995') || id.includes('571566882372')
      );
      if (foodIds.length > 0) {
        const selectedId = foodIds[index % foodIds.length];
        return this.buildUnsplashUrl(selectedId, 400, 300);
      }
    }
    return this.getFallbackImage('food', animalType, index);
  }

  /**
   * Gera URL para imagem de brinquedo
   */
  getToyImage(animalType, index) {
    const unsplashIds = IMAGE_APIS.unsplash.categories[animalType];
    if (unsplashIds && unsplashIds.length > 0) {
      const toyIds = unsplashIds.filter(id => 
        id.includes('toy') || id.includes('601758228041') || id.includes('545249390')
      );
      if (toyIds.length > 0) {
        const selectedId = toyIds[index % toyIds.length];
        return this.buildUnsplashUrl(selectedId, 400, 300);
      }
    }
    return this.getFallbackImage('toy', animalType, index);
  }

  /**
   * Gera URL para imagem de acessório
   */
  getAccessoryImage(animalType, index) {
    const accessoryIds = IMAGE_APIS.unsplash.categories.petAccessories;
    if (accessoryIds && accessoryIds.length > 0) {
      const selectedId = accessoryIds[index % accessoryIds.length];
      return this.buildUnsplashUrl(selectedId, 400, 300);
    }
    return this.getFallbackImage('accessory', animalType, index);
  }

  /**
   * Gera URL para imagem de habitação
   */
  getHousingImage(animalType, index) {
    if (animalType === 'birds') {
      const birdIds = IMAGE_APIS.unsplash.categories.birds.filter(id => 
        id.includes('cage') || id.includes('555169062')
      );
      if (birdIds.length > 0) {
        const selectedId = birdIds[index % birdIds.length];
        return this.buildUnsplashUrl(selectedId, 400, 300);
      }
    } else if (animalType === 'fish') {
      const fishIds = IMAGE_APIS.unsplash.categories.fish;
      const selectedId = fishIds[index % fishIds.length];
      return this.buildUnsplashUrl(selectedId, 400, 300);
    }
    return this.getFallbackImage('housing', animalType, index);
  }

  /**
   * Gera URL para imagem genérica do animal
   */
  getGenericAnimalImage(animalType, index) {
    const animalIds = IMAGE_APIS.unsplash.categories[animalType];
    if (animalIds && animalIds.length > 0) {
      const selectedId = animalIds[index % animalIds.length];
      return this.buildUnsplashUrl(selectedId, 400, 300);
    }
    return this.getFallbackImage('generic', animalType, index);
  }

  /**
   * Constrói URL do Unsplash
   */
  buildUnsplashUrl(photoId, width = 400, height = 300) {
    const baseUrl = IMAGE_APIS.unsplash.baseUrl;
    return `${baseUrl}/${photoId}?w=${width}&h=${height}&fit=crop&q=80`;
  }

  /**
   * Gera imagem de fallback usando Picsum ou placeholder
   */
  getFallbackImage(type, animalType, index) {
    this.apiRotation = (this.apiRotation + 1) % 2;
    
    if (this.apiRotation === 0) {
      // Usar Picsum
      const seeds = IMAGE_APIS.picsum.seeds.pets;
      const seed = seeds[index % seeds.length];
      return `${IMAGE_APIS.picsum.baseUrl}/400/300?seed=${seed}`;
    } else {
      // Usar placeholder.co
      const colors = IMAGE_APIS.placeholder.colors;
      const color = colors[index % colors.length];
      const text = this.getPlaceholderText(type, animalType);
      return `${IMAGE_APIS.placeholder.baseUrl}/400x300/${color}/FFFFFF?text=${encodeURIComponent(text)}`;
    }
  }

  /**
   * Gera texto para placeholder
   */
  getPlaceholderText(type, animalType) {
    const textMap = {
      food: `${animalType} Food`,
      toy: `${animalType} Toy`,
      accessory: `Pet Accessory`,
      housing: animalType === 'birds' ? 'Bird Cage' : 'Aquarium',
      generic: animalType.charAt(0).toUpperCase() + animalType.slice(1)
    };
    
    return textMap[type] || 'Pet Product';
  }

  /**
   * Detecta tipo de animal baseado no nome do produto
   */
  detectAnimalType(productName, category) {
    const name = productName.toLowerCase();
    
    if (name.includes('cão') || name.includes('dog') || name.includes('cachorro')) {
      return 'dogs';
    } else if (name.includes('gato') || name.includes('cat') || name.includes('felino')) {
      return 'cats';
    } else if (name.includes('pássaro') || name.includes('bird') || name.includes('canário') || name.includes('gaiola')) {
      return 'birds';
    } else if (name.includes('peixe') || name.includes('fish') || name.includes('aquário')) {
      return 'fish';
    }
    
    // Fallback baseado na categoria
    if (category === 'Gaiolas') return 'birds';
    if (category === 'Aquários') return 'fish';
    
    return 'dogs'; // Padrão
  }

  /**
   * Limpa o cache de imagens
   */
  clearCache() {
    imageCache.clear();
  }

  /**
   * Retorna estatísticas do serviço
   */
  getStats() {
    return {
      cacheSize: imageCache.size,
      requestCount: this.requestCount,
      apiRotation: this.apiRotation
    };
  }
}

// Instância singleton
const petImageService = new PetImageService();

// Exportar para uso em outros módulos
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { PetImageService, petImageService, IMAGE_APIS };
}

// Para uso em browser
if (typeof window !== 'undefined') {
  window.PetImageService = PetImageService;
  window.petImageService = petImageService;
}