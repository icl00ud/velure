import { describe, expect, it } from "vitest";
import {
  ANIMAL_ICONS,
  CATEGORY_ICONS,
  combineImagesWithFallbacks,
  detectAnimalType,
  filterValidImages,
  generateFallbackImages,
  getProductIcon,
  isValidImageUrl,
  optimizeUnsplashUrl,
} from "./image-utils";

describe("image-utils", () => {
  describe("detectAnimalType", () => {
    it("should detect dog products", () => {
      expect(detectAnimalType("Ração para cães")).toBe("dogs");
      expect(detectAnimalType("Dog food")).toBe("dogs");
      expect(detectAnimalType("Brinquedo para cachorro")).toBe("dogs");
    });

    it("should detect cat products", () => {
      expect(detectAnimalType("Ração para gatos")).toBe("cats");
      expect(detectAnimalType("Cat toy")).toBe("cats");
      expect(detectAnimalType("Arranhador felino")).toBe("cats");
    });

    it("should detect bird products", () => {
      expect(detectAnimalType("Gaiola para pássaro")).toBe("birds");
      expect(detectAnimalType("Bird food")).toBe("birds");
      expect(detectAnimalType("Comida para canário")).toBe("birds");
    });

    it("should detect fish products", () => {
      expect(detectAnimalType("Aquário para peixe")).toBe("fish");
      expect(detectAnimalType("Fish tank")).toBe("fish");
    });

    it("should detect hamster products", () => {
      expect(detectAnimalType("Gaiola para hamster")).toBe("hamsters");
    });

    it("should detect rabbit products", () => {
      expect(detectAnimalType("Comida para coelho")).toBe("rabbits");
      expect(detectAnimalType("Rabbit cage")).toBe("rabbits");
    });

    it("should return default for unknown animals", () => {
      expect(detectAnimalType("Generic pet product")).toBe("default");
    });
  });

  describe("getProductIcon", () => {
    it("should return icon for known category", () => {
      expect(getProductIcon("Any product", "Alimentação")).toBe(CATEGORY_ICONS.Alimentação);
      expect(getProductIcon("Any product", "Brinquedos")).toBe(CATEGORY_ICONS.Brinquedos);
    });

    it("should return icon based on animal type if category unknown", () => {
      expect(getProductIcon("Ração para cães")).toBe(ANIMAL_ICONS.dogs);
      expect(getProductIcon("Cat food")).toBe(ANIMAL_ICONS.cats);
    });

    it("should return default icon for unknown category and animal", () => {
      expect(getProductIcon("Generic product")).toBe(ANIMAL_ICONS.default);
    });
  });

  describe("generateFallbackImages", () => {
    it("should generate the requested number of fallback images", () => {
      const result = generateFallbackImages("Test Product", "Alimentação", 3);
      expect(result).toHaveLength(3);
    });

    it("should include category in placeholder URLs", () => {
      const result = generateFallbackImages("Test Product", "Brinquedos", 2);
      expect(result[0]).toContain("Brinquedos");
    });

    it("should default to Pet category if none provided", () => {
      const result = generateFallbackImages("Test Product", undefined, 1);
      expect(result[0]).toContain("Pet");
    });

    it("should use different colors for placeholders", () => {
      const result = generateFallbackImages("Test Product", "Test", 3);
      // All URLs should be unique due to different colors
      const uniqueUrls = new Set(result);
      expect(uniqueUrls.size).toBe(3);
    });

    it("should use Picsum when more images than colors are needed", () => {
      const result = generateFallbackImages("Test Product", "Test", 6);
      expect(result).toHaveLength(6);
      const picsumUrls = result.filter((url) => url.includes("picsum.photos"));
      expect(picsumUrls.length).toBeGreaterThan(0);
    });
  });

  describe("isValidImageUrl", () => {
    it("should validate common image extensions", () => {
      expect(isValidImageUrl("https://example.com/image.jpg")).toBe(true);
      expect(isValidImageUrl("https://example.com/image.jpeg")).toBe(true);
      expect(isValidImageUrl("https://example.com/image.png")).toBe(true);
      expect(isValidImageUrl("https://example.com/image.gif")).toBe(true);
      expect(isValidImageUrl("https://example.com/image.webp")).toBe(true);
      expect(isValidImageUrl("https://example.com/image.svg")).toBe(true);
    });

    it("should validate image URLs with query parameters", () => {
      expect(isValidImageUrl("https://example.com/image.jpg?size=large")).toBe(true);
    });

    it("should validate known image service URLs", () => {
      expect(isValidImageUrl("https://images.unsplash.com/photo-123")).toBe(true);
      expect(isValidImageUrl("https://picsum.photos/400/300")).toBe(true);
      expect(isValidImageUrl("https://placehold.co/400x300")).toBe(true);
    });

    it("should reject non-image URLs", () => {
      expect(isValidImageUrl("https://example.com/file.pdf")).toBe(false);
      expect(isValidImageUrl("https://example.com/page.html")).toBe(false);
    });

    it("should reject invalid URLs", () => {
      expect(isValidImageUrl("not-a-url")).toBe(false);
      expect(isValidImageUrl("")).toBe(false);
      expect(isValidImageUrl(null as any)).toBe(false);
      expect(isValidImageUrl(undefined as any)).toBe(false);
    });
  });

  describe("filterValidImages", () => {
    it("should filter out invalid image URLs", () => {
      const images = [
        "https://example.com/valid.jpg",
        "invalid-url",
        "https://example.com/valid.png",
        "https://example.com/file.pdf",
      ];
      const result = filterValidImages(images);
      expect(result).toHaveLength(2);
      expect(result).toContain("https://example.com/valid.jpg");
      expect(result).toContain("https://example.com/valid.png");
    });

    it("should return empty array when all URLs are invalid", () => {
      const images = ["invalid", "also-invalid"];
      const result = filterValidImages(images);
      expect(result).toHaveLength(0);
    });

    it("should handle empty input", () => {
      const result = filterValidImages([]);
      expect(result).toHaveLength(0);
    });
  });

  describe("combineImagesWithFallbacks", () => {
    it("should return only valid original images if enough are available", () => {
      const originalImages = [
        "https://example.com/1.jpg",
        "https://example.com/2.jpg",
        "https://example.com/3.jpg",
      ];
      const result = combineImagesWithFallbacks(originalImages, "Test Product", "Test", 3);
      expect(result).toHaveLength(3);
      expect(result).toEqual(originalImages);
    });

    it("should add fallbacks when not enough valid originals", () => {
      const originalImages = ["https://example.com/1.jpg"];
      const result = combineImagesWithFallbacks(originalImages, "Test Product", "Test", 3);
      expect(result).toHaveLength(3);
      expect(result[0]).toBe("https://example.com/1.jpg");
      // Next 2 should be fallbacks
      expect(result[1]).toContain("placehold.co");
    });

    it("should filter out invalid original images", () => {
      const originalImages = ["https://example.com/valid.jpg", "invalid-url", "not-image.pdf"];
      const result = combineImagesWithFallbacks(originalImages, "Test Product", "Test", 3);
      expect(result).toHaveLength(3);
      expect(result[0]).toBe("https://example.com/valid.jpg");
    });

    it("should limit to maxImages", () => {
      const originalImages = [
        "https://example.com/1.jpg",
        "https://example.com/2.jpg",
        "https://example.com/3.jpg",
        "https://example.com/4.jpg",
        "https://example.com/5.jpg",
      ];
      const result = combineImagesWithFallbacks(originalImages, "Test Product", "Test", 3);
      expect(result).toHaveLength(3);
    });
  });

  describe("optimizeUnsplashUrl", () => {
    it("should optimize Unsplash URLs with default dimensions", () => {
      const url = "https://images.unsplash.com/photo-123";
      const result = optimizeUnsplashUrl(url);
      expect(result).toContain("w=400");
      expect(result).toContain("h=300");
      expect(result).toContain("fit=crop");
      expect(result).toContain("q=80");
    });

    it("should optimize Unsplash URLs with custom dimensions", () => {
      const url = "https://images.unsplash.com/photo-123";
      const result = optimizeUnsplashUrl(url, 800, 600);
      expect(result).toContain("w=800");
      expect(result).toContain("h=600");
    });

    it("should remove existing query parameters", () => {
      const url = "https://images.unsplash.com/photo-123?w=1000&h=1000";
      const result = optimizeUnsplashUrl(url);
      expect(result).toContain("w=400");
      expect(result).toContain("h=300");
      expect(result).not.toContain("w=1000");
    });

    it("should not modify non-Unsplash URLs", () => {
      const url = "https://example.com/image.jpg";
      const result = optimizeUnsplashUrl(url);
      expect(result).toBe(url);
    });
  });
});
