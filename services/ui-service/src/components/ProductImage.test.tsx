import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor, act } from "@testing-library/react";
import { ProductImage, ProductImageWithFallback, ProductImageGallery } from "./ProductImage";

describe("ProductImage", () => {
  it("should render image with correct src and alt", () => {
    render(<ProductImage src="https://example.com/image.jpg" alt="Test Product" />);

    const img = screen.getByRole("img");
    expect(img).toBeTruthy();
    expect(img).toHaveAttribute("src", "https://example.com/image.jpg");
    expect(img).toHaveAttribute("alt", "Test Product");
  });

  it("should show loading state initially", () => {
    render(<ProductImage src="https://example.com/image.jpg" alt="Test Product" />);

    const loadingDiv = document.querySelector(".animate-pulse");
    expect(loadingDiv).toBeTruthy();
  });

  it("should show fallback icon on image error", async () => {
    render(
      <ProductImage
        src="https://example.com/broken-image.jpg"
        alt="Test Product"
        fallbackIcon="🐕"
      />
    );

    const img = screen.getByRole("img");
    // Simulate image error
    await act(async () => {
      img.dispatchEvent(new Event("error"));
    });

    await waitFor(() => {
      expect(screen.getByText("🐕")).toBeTruthy();
    });
  });

  it("should call onError callback when image fails to load", async () => {
    const onError = vi.fn();
    render(
      <ProductImage src="https://example.com/broken-image.jpg" alt="Test Product" onError={onError} />
    );

    const img = screen.getByRole("img");
    img.dispatchEvent(new Event("error"));

    await waitFor(() => {
      expect(onError).toHaveBeenCalledTimes(1);
    });
  });

  it("should call onLoad callback when image loads successfully", async () => {
    const onLoad = vi.fn();
    render(<ProductImage src="https://example.com/image.jpg" alt="Test Product" onLoad={onLoad} />);

    const img = screen.getByRole("img");
    img.dispatchEvent(new Event("load"));

    await waitFor(() => {
      expect(onLoad).toHaveBeenCalledTimes(1);
    });
  });

  it("should apply custom className", () => {
    const { container } = render(
      <ProductImage src="https://example.com/image.jpg" alt="Test Product" className="custom-class" />
    );

    expect(container.querySelector(".custom-class")).toBeTruthy();
  });

  it("should reset error state when src changes", async () => {
    const { rerender } = render(
      <ProductImage src="https://example.com/broken-image.jpg" alt="Test Product" fallbackIcon="🐕" />
    );

    const img = screen.getByRole("img");
    img.dispatchEvent(new Event("error"));

    await waitFor(() => {
      expect(screen.getByText("🐕")).toBeTruthy();
    });

    // Change src to a new image
    rerender(<ProductImage src="https://example.com/new-image.jpg" alt="Test Product" fallbackIcon="🐕" />);

    // Should show image again, not fallback
    await waitFor(() => {
      const newImg = screen.getByRole("img");
      expect(newImg.tagName).toBe("IMG");
    });
  });

  it("should have lazy loading attribute", () => {
    render(<ProductImage src="https://example.com/image.jpg" alt="Test Product" />);

    const img = screen.getByRole("img");
    expect(img).toHaveAttribute("loading", "lazy");
  });
});

describe("ProductImageWithFallback", () => {
  it("should render first image from array", () => {
    const images = ["https://example.com/image1.jpg", "https://example.com/image2.jpg"];
    render(<ProductImageWithFallback images={images} alt="Test Product" />);

    const img = screen.getByRole("img");
    expect(img).toHaveAttribute("src", images[0]);
  });

  it("should try next image when current one fails", async () => {
    const images = [
      "https://example.com/broken1.jpg",
      "https://example.com/broken2.jpg",
      "https://example.com/working.jpg",
    ];
    render(<ProductImageWithFallback images={images} alt="Test Product" />);

    // Simulate first image error
    let img = screen.getByRole("img");
    img.dispatchEvent(new Event("error"));

    await waitFor(() => {
      img = screen.getByRole("img");
      expect(img).toHaveAttribute("src", images[1]);
    });
  });

  it("should show fallback icon when all images fail", async () => {
    const images = ["https://example.com/broken1.jpg", "https://example.com/broken2.jpg"];
    render(<ProductImageWithFallback images={images} alt="Test Product" fallbackIcon="🐕" />);

    // Simulate first image failing
    let img = screen.getByRole("img");
    await act(async () => {
      img.dispatchEvent(new Event("error"));
    });

    // Wait for second image to be rendered
    await waitFor(() => {
      img = screen.getByRole("img");
      expect(img).toHaveAttribute("src", images[1]);
    });

    // Simulate second image failing
    await act(async () => {
      img.dispatchEvent(new Event("error"));
    });

    // Now fallback should be shown
    await waitFor(() => {
      expect(screen.getByText("🐕")).toBeTruthy();
    });
  });

  it("should show fallback when no images provided", () => {
    render(<ProductImageWithFallback images={[]} alt="Test Product" fallbackIcon="🐕" />);

    expect(screen.getByText("🐕")).toBeTruthy();
  });

  it("should detect product type from alt text and use appropriate icon", () => {
    render(<ProductImageWithFallback images={[]} alt="Ração para gatos" />);

    expect(screen.getByText("🐱")).toBeTruthy();
  });

  it("should detect bird products", () => {
    render(<ProductImageWithFallback images={[]} alt="Gaiola para pássaro" />);

    expect(screen.getByText("🐦")).toBeTruthy();
  });

  it("should detect fish products", () => {
    render(<ProductImageWithFallback images={[]} alt="Aquário para peixe" />);

    expect(screen.getByText("🐟")).toBeTruthy();
  });

  it("should use dog icon as default", () => {
    render(<ProductImageWithFallback images={[]} alt="Generic product" />);

    expect(screen.getByText("🐕")).toBeTruthy();
  });

  it("should reset state when images prop changes", async () => {
    const initialImages = ["https://example.com/broken.jpg"];
    const newImages = ["https://example.com/working.jpg"];

    const { rerender } = render(<ProductImageWithFallback images={initialImages} alt="Test Product" />);

    // Fail the initial image
    let img = screen.getByRole("img");
    img.dispatchEvent(new Event("error"));

    await waitFor(() => {
      expect(screen.getByText("🐕")).toBeTruthy();
    });

    // Update with new images
    rerender(<ProductImageWithFallback images={newImages} alt="Test Product" />);

    await waitFor(() => {
      const newImg = screen.getByRole("img");
      expect(newImg.tagName).toBe("IMG");
      expect(newImg).toHaveAttribute("src", newImages[0]);
    });
  });
});

describe("ProductImageGallery", () => {
  it("should render main image", () => {
    const images = ["https://example.com/image1.jpg"];
    render(<ProductImageGallery images={images} productName="Test Product" />);

    const img = screen.getByRole("img");
    expect(img).toBeTruthy();
  });

  it("should render thumbnails when multiple images provided", () => {
    const images = [
      "https://example.com/image1.jpg",
      "https://example.com/image2.jpg",
      "https://example.com/image3.jpg",
    ];
    render(<ProductImageGallery images={images} productName="Test Product" />);

    const thumbnails = screen.getAllByRole("button");
    expect(thumbnails.length).toBe(3);
  });

  it("should not render thumbnails for single image", () => {
    const images = ["https://example.com/image1.jpg"];
    render(<ProductImageGallery images={images} productName="Test Product" />);

    const thumbnails = screen.queryAllByRole("button");
    expect(thumbnails.length).toBe(0);
  });

  it("should show fallback when no images provided", () => {
    render(<ProductImageGallery images={[]} productName="Test Product" />);

    expect(screen.getByText("🐕")).toBeTruthy();
  });

  it("should limit to 3 images", () => {
    const images = [
      "https://example.com/image1.jpg",
      "https://example.com/image2.jpg",
      "https://example.com/image3.jpg",
      "https://example.com/image4.jpg",
      "https://example.com/image5.jpg",
    ];
    render(<ProductImageGallery images={images} productName="Test Product" />);

    const thumbnails = screen.getAllByRole("button");
    expect(thumbnails.length).toBe(3);
  });
});
