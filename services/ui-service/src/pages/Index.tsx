import { Bird, Cat, Dog, Fish, Heart, Loader2, Rabbit, Shield, Truck, Users } from "lucide-react";
import { Link } from "react-router-dom";
import heroImage from "@/assets/petshop-hero.png";
import Header from "@/components/Header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useCategories } from "@/hooks/use-products";
import { useEffect, useRef, useState } from "react";
import { designSystemStyles } from "@/styles/design-system";

const categoryConfig: Record<string, { name: string; icon: React.ReactNode; emoji: string }> = {
  dogs: { name: "Cães", icon: <Dog className="h-8 w-8" />, emoji: "🐕" },
  cats: { name: "Gatos", icon: <Cat className="h-8 w-8" />, emoji: "🐈" },
  birds: { name: "Pássaros", icon: <Bird className="h-8 w-8" />, emoji: "🐦" },
  fish: { name: "Peixes", icon: <Fish className="h-8 w-8" />, emoji: "🐟" },
  "small-pets": { name: "Pets Pequenos", icon: <Rabbit className="h-8 w-8" />, emoji: "🐹" },
  reptiles: { name: "Répteis", icon: null, emoji: "🦎" },
  rabbits: { name: "Coelhos", icon: <Rabbit className="h-8 w-8" />, emoji: "🐰" },
};

const Index = () => {
  const { categories, loading: loadingCategories } = useCategories();
  const [isVisible, setIsVisible] = useState(false);
  const featuresRef = useRef<HTMLElement>(null);

  useEffect(() => {
    setIsVisible(true);
  }, []);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add("animate-in");
          }
        });
      },
      { threshold: 0.1 }
    );

    const features = document.querySelectorAll(".observe-animation");
    features.forEach((feature) => observer.observe(feature));

    return () => observer.disconnect();
  }, [categories]);

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-[#F8FAF5] relative overflow-x-hidden">
        <Header />

        {/* Decorative Elements */}
        <div className="fixed top-20 right-10 w-32 h-32 rounded-full bg-[#52B788]/10 blur-3xl pointer-events-none" />
        <div className="fixed bottom-20 left-10 w-40 h-40 rounded-full bg-[#95D5B2]/10 blur-3xl pointer-events-none" />

      {/* Hero Section - Asymmetric Editorial Layout */}
      <section className="relative min-h-[90vh] flex items-center grain-texture overflow-hidden pt-20">
        {/* Decorative Circle */}
        <div className="absolute top-20 right-[15%] w-64 h-64 rounded-full border-4 border-[#52B788]/20 pointer-events-none" />

        <div className="container mx-auto px-4 lg:px-8">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            {/* Left Content */}
            <div className={`relative z-10 ${isVisible ? 'hero-enter active' : 'hero-enter'}`}>
              <div className="inline-block mb-6 px-6 py-2 bg-[#95D5B2]/20 rounded-full">
                <span className="font-body text-[#2D6A4F] text-sm font-medium tracking-wider uppercase">
                  Premium Pet Care
                </span>
              </div>

              <h1 className="font-display text-6xl lg:text-8xl font-bold text-[#1B4332] mb-6 leading-[0.95] text-shadow-warm">
                Tudo que seu
                <span className="block text-[#52B788] italic">pet precisa</span>
              </h1>

              <div className="w-20 h-1 bg-gradient-to-r from-[#52B788] to-[#A7C957] mb-8" />

              <p className="font-body text-xl text-[#2D6A4F] mb-10 leading-relaxed max-w-lg">
                De ração premium a camas aconchegantes, brinquedos a cuidados de saúde.
                Cuidamos de cada detalhe para a felicidade dos seus companheiros.
              </p>

              <div className="flex flex-wrap gap-4">
                <Link to="/products/dogs">
                  <button className="btn-primary-custom font-body px-10 py-4 rounded-full text-white font-semibold text-lg">
                    Explorar Produtos
                  </button>
                </Link>
                <Link to="/contact">
                  <button className="font-body px-10 py-4 rounded-full border-3 border-[#1B4332] text-[#1B4332] font-semibold text-lg hover:bg-[#1B4332] hover:text-white transition-all duration-300">
                    Fale Conosco
                  </button>
                </Link>
              </div>
            </div>

            {/* Right Image - Asymmetric */}
            <div
              className={`relative lg:absolute lg:right-0 lg:top-20 lg:w-[50%] h-[500px] lg:h-[600px] ${isVisible ? 'hero-enter active' : 'hero-enter'}`}
              style={{ animationDelay: '0.2s' }}
            >
              <div className="relative h-full rounded-3xl overflow-hidden shadow-2xl transform lg:rotate-2">
                <img
                  src={heroImage}
                  alt="Pets felizes"
                  className="w-full h-full object-cover"
                />
                <div className="absolute inset-0 bg-gradient-to-t from-[#1B4332]/30 to-transparent" />
              </div>

              {/* Decorative Element */}
              <div className="absolute -bottom-6 -left-6 w-32 h-32 bg-[#A7C957] rounded-3xl -z-10 transform rotate-12" />
              <div className="absolute -top-6 -right-6 w-24 h-24 bg-[#95D5B2] rounded-full -z-10" />
            </div>
          </div>
        </div>
      </section>

      {/* Features Section - Overlapping Cards */}
      <section ref={featuresRef} className="py-32 relative">
        <div className="container mx-auto px-4 lg:px-8">
          {/* Section Header */}
          <div className="max-w-3xl mb-20 observe-animation">
            <span className="font-body text-[#52B788] font-semibold text-sm tracking-widest uppercase mb-4 block">
              Por Que Escolher Velure
            </span>
            <h2 className="font-display text-5xl lg:text-6xl font-bold text-[#1B4332] mb-6 leading-tight">
              Parceiros no cuidado com pets
            </h2>
            <div className="w-16 h-1 bg-[#52B788]" />
          </div>

          {/* Features Grid - Staggered */}
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            {[
              {
                icon: Heart,
                title: "Qualidade Premium",
                description: "Apenas os melhores produtos de marcas confiáveis para seus pets amados",
                color: "#52B788",
                delay: "0s"
              },
              {
                icon: Shield,
                title: "Saúde Garantida",
                description: "Todos os produtos são aprovados por veterinários e vêm com garantias",
                color: "#95D5B2",
                delay: "0.1s"
              },
              {
                icon: Truck,
                title: "Entrega Rápida",
                description: "Frete grátis em pedidos acima de R$ 250. Entrega no mesmo dia disponível",
                color: "#A7C957",
                delay: "0.2s"
              },
              {
                icon: Users,
                title: "Suporte Especializado",
                description: "Nossos especialistas em cuidados com pets estão aqui para ajudá-lo",
                color: "#52B788",
                delay: "0.3s"
              }
            ].map((feature, index) => {
              const Icon = feature.icon;
              return (
                <div
                  key={index}
                  className="observe-animation card-hover"
                  style={{ animationDelay: feature.delay }}
                >
                  <div className="relative bg-white rounded-2xl p-8 shadow-lg hover:shadow-2xl border-3 border-transparent hover:border-[#1B4332]/10 h-full">
                    {/* Icon */}
                    <div
                      className="w-16 h-16 rounded-2xl mb-6 flex items-center justify-center transform -rotate-6"
                      style={{ backgroundColor: `${feature.color}20` }}
                    >
                      <Icon className="h-8 w-8" style={{ color: feature.color }} />
                    </div>

                    <h3 className="font-display text-2xl font-bold text-[#1B4332] mb-4">
                      {feature.title}
                    </h3>

                    <p className="font-body text-[#2D6A4F] leading-relaxed">
                      {feature.description}
                    </p>

                    {/* Decorative Corner */}
                    <div
                      className="absolute top-0 right-0 w-20 h-20 rounded-bl-full opacity-5"
                      style={{ backgroundColor: feature.color }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* Categories Section - Bento Grid */}
      <section className="py-32 bg-gradient-to-b from-white to-[#F8FAF5]">
        <div className="container mx-auto px-4 lg:px-8">
          {/* Section Header */}
          <div className="text-center mb-20 observe-animation">
            <span className="font-body text-[#95D5B2] font-semibold text-sm tracking-widest uppercase mb-4 block">
              Categorias
            </span>
            <h2 className="font-display text-5xl lg:text-6xl font-bold text-[#1B4332] mb-6">
              Compre por pet
            </h2>
            <p className="font-body text-xl text-[#2D6A4F] max-w-2xl mx-auto">
              Encontre tudo que seu pet específico precisa
            </p>
          </div>

          {loadingCategories ? (
            <div className="flex justify-center items-center py-20">
              <Loader2 className="h-12 w-12 animate-spin text-[#52B788]" />
              <span className="ml-4 font-body text-[#2D6A4F] text-lg">Carregando categorias...</span>
            </div>
          ) : (
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-6">
              {categories.map((category, index) => {
                const config = categoryConfig[category.toLowerCase()] || {
                  name: category,
                  icon: null,
                  emoji: "🐾",
                };
                return (
                  <Link key={category} to={`/products/${category}`}>
                    <div
                      className="category-card observe-animation bg-white rounded-3xl p-8 text-center h-full shadow-lg group relative overflow-hidden"
                      style={{ animationDelay: `${index * 0.05}s` }}
                    >
                      {/* Background Gradient */}
                      <div className="absolute inset-0 bg-gradient-to-br from-[#52B788]/5 to-[#95D5B2]/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500" />

                      {/* Content */}
                      <div className="relative z-10">
                        <div className="mb-6 flex justify-center transform group-hover:scale-110 group-hover:rotate-12 transition-all duration-500">
                          {config.icon ? (
                            <div className="bg-gradient-to-br from-[#52B788] to-[#40916C] rounded-2xl p-4 text-white shadow-lg">
                              {config.icon}
                            </div>
                          ) : (
                            <span className="text-6xl filter drop-shadow-lg">
                              {config.emoji}
                            </span>
                          )}
                        </div>

                        <h3 className="font-display text-xl font-bold text-[#1B4332] group-hover:text-[#52B788] transition-colors duration-300">
                          {config.name}
                        </h3>
                      </div>

                      {/* Decorative Circle */}
                      <div className="absolute -bottom-10 -right-10 w-32 h-32 rounded-full bg-[#A7C957]/20 transform group-hover:scale-150 transition-transform duration-700" />
                    </div>
                  </Link>
                );
              })}
            </div>
          )}
        </div>
      </section>

      {/* CTA Section - Diagonal Split */}
      <section className="relative py-32 overflow-hidden">
        {/* Background with Diagonal */}
        <div className="absolute inset-0 bg-gradient-to-br from-[#1B4332] via-[#2D6A4F] to-[#1B4332] diagonal-split" />

        {/* Decorative Elements */}
        <div className="absolute top-10 left-[10%] w-40 h-40 rounded-full border-4 border-white/10" />
        <div className="absolute bottom-10 right-[10%] w-32 h-32 rounded-full bg-[#52B788]/20" />

        <div className="container mx-auto px-4 lg:px-8 relative z-10">
          <div className="max-w-4xl mx-auto text-center observe-animation">
            <div className="inline-block mb-6 px-6 py-2 bg-white/10 backdrop-blur-sm rounded-full">
              <span className="font-body text-white/90 text-sm font-medium tracking-wider uppercase">
                Junte-se a Nós
              </span>
            </div>

            <h2 className="font-display text-5xl lg:text-7xl font-bold text-white mb-8 leading-tight text-shadow-warm">
              Pronto para mimar
              <span className="block text-[#A7C957] italic">seu pet?</span>
            </h2>

            <p className="font-body text-xl text-white/80 mb-12 leading-relaxed max-w-2xl mx-auto">
              Junte-se a milhares de tutores felizes que confiam em nós para a felicidade e saúde de seus pets
            </p>

            <Link to="/products/dogs">
              <button className="font-body px-12 py-5 rounded-full bg-white text-[#1B4332] font-bold text-lg shadow-2xl hover:scale-105 hover:shadow-[#A7C957]/50 transition-all duration-300">
                Começar a Comprar
              </button>
            </Link>

            {/* Stats */}
            <div className="grid grid-cols-3 gap-8 mt-20 max-w-3xl mx-auto">
              {[
                { number: "10k+", label: "Tutores Felizes" },
                { number: "500+", label: "Produtos" },
                { number: "98%", label: "Satisfação" }
              ].map((stat, index) => (
                <div key={index} className="text-center observe-animation" style={{ animationDelay: `${index * 0.1}s` }}>
                  <div className="font-display text-4xl lg:text-5xl font-bold text-[#A7C957] mb-2">
                    {stat.number}
                  </div>
                  <div className="font-body text-white/70 text-sm uppercase tracking-wider">
                    {stat.label}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>
      </div>
    </>
  );
};

export default Index;
