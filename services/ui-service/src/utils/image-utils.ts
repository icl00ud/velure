/**
 * Utilitários para URLs de imagem e fallbacks
 */

// Configuração das APIs de imagem
const IMAGE_CONFIG = {
  unsplash: {
    baseUrl: "https://images.unsplash.com",
    defaultParams: "w=400&h=300&fit=crop&q=80",
  },
  picsum: {
    baseUrl: "https://picsum.photos",
    size: "400/300",
  },
  placeholder: {
    baseUrl: "https://placehold.co",
    size: "400x300",
    defaultBg: "8B4513",
    defaultText: "FFFFFF",
  },
};

// Mapeamento de categorias para ícones
const CATEGORY_ICONS = {
  Alimentação: "🍽️",
  Brinquedos: "🎾",
  Acessórios: "🎀",
  "Camas e Descanso": "🛏️",
  Higiene: "🧼",
  Transporte: "🧳",
  Gaiolas: "🏠",
  Aquários: "🐠",
  Suplementos: "💊",
  Roupas: "👕",
  default: "🐕",
};

// Mapeamento de tipos de animal para ícones
const ANIMAL_ICONS = {
  dogs: "🐕",
  cats: "🐱",
  birds: "🐦",
  fish: "🐟",
  hamsters: "🐹",
  rabbits: "🐰",
  default: "🐾",
};

/**
 * Detecta o tipo de animal baseado no nome do produto
 */
export function detectAnimalType(productName: string): keyof typeof ANIMAL_ICONS {
  const name = productName.toLowerCase();

  if (name.includes("cão") || name.includes("cães") || name.includes("dog") || name.includes("cachorro")) {
    return "dogs";
  }
  if (name.includes("gato") || name.includes("cat") || name.includes("felino")) {
    return "cats";
  }
  if (
    name.includes("pássaro") ||
    name.includes("bird") ||
    name.includes("canário") ||
    name.includes("papagaio")
  ) {
    return "birds";
  }
  if (name.includes("peixe") || name.includes("fish") || name.includes("aquário")) {
    return "fish";
  }
  if (name.includes("hamster")) {
    return "hamsters";
  }
  if (name.includes("coelho") || name.includes("rabbit")) {
    return "rabbits";
  }

  return "default";
}

/**
 * Gera ícone apropriado baseado no produto
 */
export function getProductIcon(productName: string, category?: string): string {
  // Primeiro tenta por categoria
  if (category && CATEGORY_ICONS[category as keyof typeof CATEGORY_ICONS]) {
    return CATEGORY_ICONS[category as keyof typeof CATEGORY_ICONS];
  }

  // Depois por tipo de animal
  const animalType = detectAnimalType(productName);
  return ANIMAL_ICONS[animalType];
}

/**
 * Gera URLs de fallback para imagens
 */
export function generateFallbackImages(
  productName: string,
  category?: string,
  count: number = 3
): string[] {
  const fallbacks: string[] = [];
  const icon = getProductIcon(productName, category);
  const encodedText = encodeURIComponent(`${category || "Pet"} Product`);

  // Adicionar diferentes variações de placeholder
  const colors = ["8B4513", "228B22", "4169E1", "FF6347", "9370DB"];

  for (let i = 0; i < Math.min(count, colors.length); i++) {
    const color = colors[i];
    fallbacks.push(
      `${IMAGE_CONFIG.placeholder.baseUrl}/${IMAGE_CONFIG.placeholder.size}/${color}/${IMAGE_CONFIG.placeholder.defaultText}?text=${encodedText}`
    );
  }

  // Se precisar de mais imagens, usar Picsum com diferentes seeds
  while (fallbacks.length < count) {
    const seed = fallbacks.length + 1;
    fallbacks.push(`${IMAGE_CONFIG.picsum.baseUrl}/${IMAGE_CONFIG.picsum.size}?seed=${seed}`);
  }

  return fallbacks;
}

/**
 * Valida se uma URL de imagem é válida
 */
export function isValidImageUrl(url: string): boolean {
  if (!url || typeof url !== "string") return false;

  try {
    new URL(url);
    return (
      /\.(jpg|jpeg|png|gif|webp|svg)(\?.*)?$/i.test(url) ||
      url.includes("unsplash.com") ||
      url.includes("picsum.photos") ||
      url.includes("placehold.co")
    );
  } catch {
    return false;
  }
}

/**
 * Filtra URLs de imagem válidas de um array
 */
export function filterValidImages(images: string[]): string[] {
  return images.filter(isValidImageUrl);
}

/**
 * Combina imagens originais com fallbacks
 */
export function combineImagesWithFallbacks(
  originalImages: string[],
  productName: string,
  category?: string,
  maxImages: number = 3
): string[] {
  const validOriginals = filterValidImages(originalImages);

  if (validOriginals.length >= maxImages) {
    return validOriginals.slice(0, maxImages);
  }

  const fallbacks = generateFallbackImages(
    productName,
    category,
    maxImages - validOriginals.length
  );

  return [...validOriginals, ...fallbacks].slice(0, maxImages);
}

/**
 * Gera URL otimizada do Unsplash
 */
export function optimizeUnsplashUrl(
  url: string,
  width: number = 400,
  height: number = 300
): string {
  if (!url.includes("unsplash.com")) return url;

  // Remove parâmetros existentes e adiciona novos
  const baseUrl = url.split("?")[0];
  return `${baseUrl}?${IMAGE_CONFIG.unsplash.defaultParams}&w=${width}&h=${height}`;
}

/**
 * Hook para gerenciar carregamento de imagens
 */
export function useImageLoader() {
  const loadImage = (src: string): Promise<boolean> => {
    return new Promise((resolve) => {
      const img = new Image();
      img.onload = () => resolve(true);
      img.onerror = () => resolve(false);
      img.src = src;
    });
  };

  const loadImages = async (urls: string[]): Promise<string[]> => {
    const results = await Promise.all(
      urls.map(async (url) => {
        const success = await loadImage(url);
        return success ? url : null;
      })
    );

    return results.filter(Boolean) as string[];
  };

  return { loadImage, loadImages };
}

/**
 * Configuração de lazy loading para imagens
 */
export const IMAGE_LAZY_CONFIG = {
  rootMargin: "50px",
  threshold: 0.1,
};

export { IMAGE_CONFIG, CATEGORY_ICONS, ANIMAL_ICONS };
