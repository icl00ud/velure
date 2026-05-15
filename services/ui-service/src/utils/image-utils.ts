/**
 * Utilities for image URLs and fallbacks
 */

// Image API configuration
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

// Category → icon mapping
const CATEGORY_ICONS = {
  Food: "🍽️",
  Toys: "🎾",
  Accessories: "🎀",
  "Beds & Rest": "🛏️",
  Hygiene: "🧼",
  Transport: "🧳",
  Cages: "🏠",
  Aquariums: "🐠",
  Supplements: "💊",
  Apparel: "👕",
  default: "🐕",
};

// Animal type → icon mapping
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
 * Detects the animal type from the product name.
 */
export function detectAnimalType(productName: string): keyof typeof ANIMAL_ICONS {
  const name = productName.toLowerCase();

  if (
    name.includes("dog") ||
    name.includes("puppy") ||
    name.includes("canine")
  ) {
    return "dogs";
  }
  if (name.includes("cat") || name.includes("kitten") || name.includes("feline")) {
    return "cats";
  }
  if (
    name.includes("bird") ||
    name.includes("canary") ||
    name.includes("parrot")
  ) {
    return "birds";
  }
  if (name.includes("fish") || name.includes("aquarium")) {
    return "fish";
  }
  if (name.includes("hamster")) {
    return "hamsters";
  }
  if (name.includes("rabbit") || name.includes("bunny")) {
    return "rabbits";
  }

  return "default";
}

/**
 * Returns an icon for the given product.
 */
export function getProductIcon(productName: string, category?: string): string {
  // Try by category first
  if (category && CATEGORY_ICONS[category as keyof typeof CATEGORY_ICONS]) {
    return CATEGORY_ICONS[category as keyof typeof CATEGORY_ICONS];
  }

  // Then by animal type
  const animalType = detectAnimalType(productName);
  return ANIMAL_ICONS[animalType];
}

/**
 * Generates fallback image URLs.
 */
export function generateFallbackImages(
  productName: string,
  category?: string,
  count: number = 3
): string[] {
  const fallbacks: string[] = [];
  const _icon = getProductIcon(productName, category);
  const encodedText = encodeURIComponent(`${category || "Pet"} Product`);

  // Different placeholder variants
  const colors = ["8B4513", "228B22", "4169E1", "FF6347", "9370DB"];

  for (let i = 0; i < Math.min(count, colors.length); i++) {
    const color = colors[i];
    fallbacks.push(
      `${IMAGE_CONFIG.placeholder.baseUrl}/${IMAGE_CONFIG.placeholder.size}/${color}/${IMAGE_CONFIG.placeholder.defaultText}?text=${encodedText}`
    );
  }

  // If more images are needed, fall back to Picsum with different seeds
  while (fallbacks.length < count) {
    const seed = fallbacks.length + 1;
    fallbacks.push(`${IMAGE_CONFIG.picsum.baseUrl}/${IMAGE_CONFIG.picsum.size}?seed=${seed}`);
  }

  return fallbacks;
}

/**
 * Returns whether an image URL is valid.
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
 * Filters valid image URLs from an array.
 */
export function filterValidImages(images: string[]): string[] {
  return images.filter(isValidImageUrl);
}

/**
 * Combines original images with fallbacks.
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
 * Generates an optimized Unsplash URL.
 */
export function optimizeUnsplashUrl(
  url: string,
  width: number = 400,
  height: number = 300
): string {
  if (!url.includes("unsplash.com")) return url;

  // Strip existing parameters and add new ones
  const baseUrl = url.split("?")[0];
  return `${baseUrl}?${IMAGE_CONFIG.unsplash.defaultParams}&w=${width}&h=${height}`;
}

/**
 * Hook to manage image loading.
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
 * Lazy loading configuration for images.
 */
export const IMAGE_LAZY_CONFIG = {
  rootMargin: "50px",
  threshold: 0.1,
};

export { IMAGE_CONFIG, CATEGORY_ICONS, ANIMAL_ICONS };
