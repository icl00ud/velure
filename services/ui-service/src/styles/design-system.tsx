// Velure Design System
// Warm Editorial Design - Shared styles, colors, fonts, and animations

import React from 'react';

export const designSystemStyles = `
  @import url('https://fonts.googleapis.com/css2?family=Playfair+Display:wght@500;600;700;800;900&family=Outfit:wght@300;400;500;600;700&display=swap');

  .font-display { font-family: 'Playfair Display', serif; }
  .font-body { font-family: 'Outfit', sans-serif; }

  /* Color Variables */
  :root {
    --color-terracotta: #D97757;
    --color-terracotta-dark: #C56647;
    --color-sage: #8B9A7E;
    --color-sage-dark: #5A6751;
    --color-warm-yellow: #F4C430;
    --color-cream: #FAF7F2;
    --color-charcoal: #2D3319;
    --color-charcoal-light: #3D4428;
  }

  /* Grain Texture */
  .grain-texture {
    position: relative;
  }

  .grain-texture::before {
    content: '';
    position: absolute;
    inset: 0;
    background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 400 400' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)' opacity='0.03'/%3E%3C/svg%3E");
    pointer-events: none;
    z-index: 1;
  }

  /* Page Animations */
  .page-enter {
    animation: pageSlideUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
    opacity: 0;
    transform: translateY(30px);
  }

  .page-enter.active {
    opacity: 1;
    transform: translateY(0);
  }

  @keyframes pageSlideUp {
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* Hero Animations */
  .hero-enter {
    animation: heroSlideUp 1.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
    opacity: 0;
    transform: translateY(40px);
  }

  .hero-enter.active {
    opacity: 1;
    transform: translateY(0);
  }

  @keyframes heroSlideUp {
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* Scroll Animations */
  .observe-animation {
    opacity: 0;
    transform: translateY(30px);
    transition: all 0.8s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .observe-animation.animate-in {
    opacity: 1;
    transform: translateY(0);
  }

  /* Card Hover Effects */
  .card-hover {
    transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .card-hover:hover {
    transform: translateY(-8px) rotate(-1deg);
  }

  .card-hover-subtle {
    transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .card-hover-subtle:hover {
    transform: translateY(-4px);
  }

  /* Product Card Styles */
  .product-card {
    transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
    border: 3px solid transparent;
  }

  .product-card:hover {
    transform: translateY(-8px);
    border-color: var(--color-terracotta);
    box-shadow: 0 20px 60px rgba(217, 119, 87, 0.25);
  }

  /* Category Card Styles */
  .category-card {
    transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
    border: 3px solid transparent;
  }

  .category-card:hover {
    transform: scale(1.05) rotate(2deg);
    border-color: var(--color-terracotta);
    box-shadow: 0 20px 60px rgba(217, 119, 87, 0.25);
  }

  /* Decorative Border Effect */
  .decorative-border {
    position: relative;
  }

  .decorative-border::after {
    content: '';
    position: absolute;
    top: -10px;
    left: -10px;
    right: -10px;
    bottom: -10px;
    border: 2px solid var(--color-terracotta);
    border-radius: inherit;
    opacity: 0;
    transition: opacity 0.3s ease;
  }

  .decorative-border:hover::after {
    opacity: 0.3;
  }

  /* Text Shadows */
  .text-shadow-warm {
    text-shadow: 0 2px 20px rgba(217, 119, 87, 0.15);
  }

  .text-shadow-subtle {
    text-shadow: 0 1px 10px rgba(45, 51, 25, 0.1);
  }

  /* Diagonal Effects */
  .diagonal-split {
    clip-path: polygon(0 0, 100% 5%, 100% 100%, 0% 100%);
  }

  .diagonal-split-reverse {
    clip-path: polygon(0 5%, 100% 0, 100% 100%, 0% 100%);
  }

  /* Button Styles */
  .btn-primary-custom {
    background: linear-gradient(135deg, var(--color-terracotta) 0%, var(--color-terracotta-dark) 100%);
    box-shadow: 0 10px 30px rgba(217, 119, 87, 0.3);
    transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .btn-primary-custom:hover {
    transform: translateY(-2px);
    box-shadow: 0 15px 40px rgba(217, 119, 87, 0.4);
  }

  .btn-secondary-custom {
    border: 3px solid var(--color-charcoal);
    color: var(--color-charcoal);
    transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .btn-secondary-custom:hover {
    background: var(--color-charcoal);
    color: white;
    transform: translateY(-2px);
  }

  /* Badge Styles */
  .badge-warm {
    background: rgba(139, 154, 126, 0.2);
    color: var(--color-sage-dark);
    border-radius: 9999px;
    padding: 0.5rem 1.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }

  /* Section Dividers */
  .section-divider {
    width: 4rem;
    height: 0.25rem;
    background: linear-gradient(to right, var(--color-terracotta), var(--color-warm-yellow));
  }

  .section-divider-small {
    width: 4rem;
    height: 0.25rem;
    background: var(--color-terracotta);
  }

  /* Loading Spinner */
  .spinner-warm {
    color: var(--color-terracotta);
  }

  /* Form Inputs */
  .input-warm {
    border: 2px solid rgba(45, 51, 25, 0.1);
    border-radius: 1rem;
    padding: 1rem 1.5rem;
    transition: all 0.3s ease;
    background: white;
  }

  .input-warm:focus {
    border-color: var(--color-terracotta);
    outline: none;
    box-shadow: 0 0 0 3px rgba(217, 119, 87, 0.1);
  }

  /* Decorative Elements */
  .decorative-circle {
    border-radius: 9999px;
    border: 4px solid rgba(217, 119, 87, 0.2);
  }

  .decorative-blob {
    border-radius: 30% 70% 70% 30% / 30% 30% 70% 70%;
    background: linear-gradient(135deg, var(--color-terracotta) 0%, var(--color-warm-yellow) 100%);
    opacity: 0.1;
  }

  /* Backdrop Blur */
  .backdrop-warm {
    backdrop-filter: blur(10px);
    background: rgba(250, 247, 242, 0.8);
  }

  /* Image Overlays */
  .image-overlay-dark {
    background: linear-gradient(to top, rgba(45, 51, 25, 0.3), transparent);
  }

  .image-overlay-warm {
    background: linear-gradient(to top, rgba(217, 119, 87, 0.3), transparent);
  }

  /* Responsive Typography */
  @media (max-width: 768px) {
    .font-display {
      font-size: clamp(2rem, 5vw, 3rem);
    }
  }

  /* Selection Color */
  ::selection {
    background: rgba(217, 119, 87, 0.3);
    color: var(--color-charcoal);
  }

  /* Smooth Scrolling */
  html {
    scroll-behavior: smooth;
  }

  /* Custom Scrollbar */
  ::-webkit-scrollbar {
    width: 10px;
  }

  ::-webkit-scrollbar-track {
    background: var(--color-cream);
  }

  ::-webkit-scrollbar-thumb {
    background: var(--color-terracotta);
    border-radius: 5px;
  }

  ::-webkit-scrollbar-thumb:hover {
    background: var(--color-terracotta-dark);
  }
`;

// Intersection Observer Hook for Scroll Animations
export const useScrollAnimation = (callback?: () => void) => {
  const observerRef = React.useRef<IntersectionObserver | null>(null);

  React.useEffect(() => {
    observerRef.current = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add("animate-in");
            if (callback) callback();
          }
        });
      },
      { threshold: 0.1 }
    );

    const elements = document.querySelectorAll(".observe-animation");
    elements.forEach((element) => observerRef.current?.observe(element));

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, [callback]);
};

// Color Constants
export const colors = {
  terracotta: '#D97757',
  terracottaDark: '#C56647',
  sage: '#8B9A7E',
  sageDark: '#5A6751',
  warmYellow: '#F4C430',
  cream: '#FAF7F2',
  charcoal: '#2D3319',
  charcoalLight: '#3D4428',
};

// Animation Delays
export const staggerDelays = (index: number, baseDelay = 0.05) => ({
  animationDelay: `${index * baseDelay}s`,
});
