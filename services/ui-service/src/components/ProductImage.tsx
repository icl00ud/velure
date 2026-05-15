import type React from "react";
import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";

interface ProductImageProps {
  src: string;
  alt: string;
  className?: string;
  fallbackIcon?: string;
  onError?: () => void;
  onLoad?: () => void;
}

interface ProductImageWithFallbackProps {
  images: string[];
  alt: string;
  className?: string;
  fallbackIcon?: string;
  priority?: boolean;
}

/**
 * Componente de imagem com fallback automático para produtos
 * Tenta múltiplas URLs antes de mostrar um fallback
 */
const ProductImage: React.FC<ProductImageProps> = ({
  src,
  alt,
  className,
  fallbackIcon = "🐕",
  onError,
  onLoad,
}) => {
  const [imageError, setImageError] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const handleImageError = () => {
    setImageError(true);
    setIsLoading(false);
    onError?.();
  };

  const handleImageLoad = () => {
    setIsLoading(false);
    onLoad?.();
  };

  // biome-ignore lint/correctness/useExhaustiveDependencies: Reset state when the image source changes.
  useEffect(() => {
    // Reset states when src changes
    setImageError(false);
    setIsLoading(true);
  }, [src]);

  if (imageError) {
    return (
      <div
        role="img"
        className={cn(
          "bg-muted rounded-lg flex items-center justify-center text-4xl transition-colors",
          "hover:bg-muted/80",
          className
        )}
        aria-label={alt}
      >
        {fallbackIcon}
      </div>
    );
  }

  return (
    <div className={cn("relative overflow-hidden rounded-lg", className)}>
      {isLoading && <div className="absolute inset-0 bg-muted animate-pulse rounded-lg" />}
      <img
        src={src}
        alt={alt}
        className={cn(
          "w-full h-full object-cover transition-opacity duration-300",
          isLoading ? "opacity-0" : "opacity-100"
        )}
        onError={handleImageError}
        onLoad={handleImageLoad}
        loading="lazy"
      />
    </div>
  );
};

/**
 * Componente avançado que tenta múltiplas imagens antes do fallback
 */
const ProductImageWithFallback: React.FC<ProductImageWithFallbackProps> = ({
  images,
  alt,
  className,
  fallbackIcon = "🐕",
}) => {
  const [currentImageIndex, setCurrentImageIndex] = useState(0);
  const [allImagesFailed, setAllImagesFailed] = useState(false);

  // Determinar o ícone baseado no tipo de produto
  const getProductIcon = (productName: string): string => {
    const name = productName.toLowerCase();

    if (
      name.includes("cão") ||
      name.includes("cães") ||
      name.includes("dog") ||
      name.includes("cachorro")
    ) {
      return "🐕";
    } else if (name.includes("gato") || name.includes("cat") || name.includes("felino")) {
      return "🐱";
    } else if (name.includes("pássaro") || name.includes("bird") || name.includes("canário")) {
      return "🐦";
    } else if (name.includes("peixe") || name.includes("fish") || name.includes("aquário")) {
      return "🐟";
    } else if (name.includes("hamster")) {
      return "🐹";
    } else if (name.includes("coelho") || name.includes("rabbit")) {
      return "🐰";
    }

    return "🐕"; // Padrão para cães
  };

  const iconToUse = getProductIcon(alt) || fallbackIcon;

  const handleImageError = () => {
    const nextIndex = currentImageIndex + 1;

    if (nextIndex < images.length) {
      setCurrentImageIndex(nextIndex);
    } else {
      setAllImagesFailed(true);
    }
  };

  const handleImageLoad = () => {
    // Imagem carregou com sucesso, não fazer nada
  };

  // biome-ignore lint/correctness/useExhaustiveDependencies: Reset state when the image list changes.
  useEffect(() => {
    // Reset quando as imagens mudam
    setCurrentImageIndex(0);
    setAllImagesFailed(false);
  }, [images]);

  if (!images || images.length === 0 || allImagesFailed) {
    return (
      <div
        role="img"
        className={cn(
          "bg-muted rounded-lg flex items-center justify-center text-4xl transition-colors",
          "hover:bg-muted/80",
          className
        )}
        aria-label={alt}
      >
        {iconToUse}
      </div>
    );
  }

  const currentImage = images[currentImageIndex];

  return (
    <ProductImage
      src={currentImage}
      alt={alt}
      className={className}
      fallbackIcon={iconToUse}
      onError={handleImageError}
      onLoad={handleImageLoad}
    />
  );
};

/**
 * Hook para gerenciar carregamento de imagens com cache
 */
const useImagePreloader = (images: string[]) => {
  const [loadedImages, setLoadedImages] = useState<Set<string>>(new Set());
  const [failedImages, setFailedImages] = useState<Set<string>>(new Set());

  useEffect(() => {
    const preloadImage = (src: string): Promise<void> => {
      return new Promise((resolve) => {
        const img = new Image();

        img.onload = () => {
          setLoadedImages((prev) => new Set(prev).add(src));
          resolve();
        };

        img.onerror = () => {
          setFailedImages((prev) => new Set(prev).add(src));
          resolve();
        };

        img.src = src;
      });
    };

    // Precarregar todas as imagens
    Promise.all(images.map(preloadImage));
  }, [images]);

  return {
    isLoaded: (src: string) => loadedImages.has(src),
    hasFailed: (src: string) => failedImages.has(src),
    loadedCount: loadedImages.size,
    failedCount: failedImages.size,
  };
};

/**
 * Componente de galeria de imagens do produto
 */
interface ProductImageGalleryProps {
  images: string[];
  productName: string;
  className?: string;
}

const ProductImageGallery: React.FC<ProductImageGalleryProps> = ({
  images,
  productName,
  className,
}) => {
  const [selectedImage, setSelectedImage] = useState(0);
  const { hasFailed } = useImagePreloader(images);

  // Filtrar imagens que falharam
  const validImages = images.filter((_, index) => !hasFailed(images[index]) && index < 3);

  if (validImages.length === 0) {
    return <ProductImageWithFallback images={[]} alt={productName} className={className} />;
  }

  return (
    <div className={cn("space-y-2", className)}>
      {/* Imagem principal */}
      <div className="aspect-square">
        <ProductImageWithFallback
          images={[validImages[selectedImage]]}
          alt={productName}
          className="w-full h-full"
          priority={selectedImage === 0}
        />
      </div>

      {/* Thumbnails */}
      {validImages.length > 1 && (
        <div className="flex gap-2">
          {validImages.map((image, index) => (
            <button
              key={image}
              type="button"
              onClick={() => setSelectedImage(index)}
              className={cn(
                "w-16 h-16 border-2 rounded-lg overflow-hidden transition-colors",
                selectedImage === index ? "border-primary" : "border-muted hover:border-primary/50"
              )}
            >
              <ProductImageWithFallback
                images={[image]}
                alt={`${productName} - Imagem ${index + 1}`}
                className="w-full h-full"
              />
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export { ProductImage, ProductImageWithFallback, ProductImageGallery, useImagePreloader };
