import { Bird, Cat, Dog, Fish, Heart, Loader2, Rabbit, Shield, Truck, Users } from "lucide-react";
import { Link } from "react-router-dom";
import heroImage from "@/assets/petshop-hero.jpg";
import Header from "@/components/Header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useCategories } from "@/hooks/use-products";

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

  return (
    <div className="min-h-screen bg-background">
      <Header />

      {/* Hero Section */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0">
          <img
            src={heroImage}
            alt="Pets felizes em nossa loja"
            className="w-full h-full object-cover"
          />
          <div className="absolute inset-0 bg-gradient-to-r from-background/90 via-background/50 to-transparent"></div>
        </div>

        <div className="relative container mx-auto px-4 py-24 lg:py-32">
          <div className="max-w-2xl">
            <h1 className="text-5xl lg:text-6xl font-bold text-foreground mb-6">
              Tudo que seu
              <span className="text-primary block">pet precisa</span>
            </h1>
            <p className="text-xl text-muted-foreground mb-8 leading-relaxed">
              De ração premium a camas aconchegantes, brinquedos a cuidados de saúde - temos tudo
              para manter seus amigos peludos, emplumados e nadadeiros felizes e saudáveis.
            </p>
            <div className="flex flex-col sm:flex-row gap-4">
              <Button
                size="lg"
                className="bg-gradient-primary hover:opacity-90 text-primary-foreground px-8"
                asChild
              >
                <Link to="/products/dogs">Comprar agora</Link>
              </Button>
              <Button
                variant="outline"
                size="lg"
                className="border-primary text-primary hover:bg-primary hover:text-primary-foreground"
                asChild
              >
                <Link to="/contact">Saiba mais</Link>
              </Button>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-muted/30">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-foreground mb-4">
              Por que tutores nos escolhem
            </h2>
            <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
              Somos mais do que uma pet shop - somos seu parceiro no cuidado com pets
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            <Card className="text-center shadow-soft hover:shadow-primary transition-all duration-300">
              <CardHeader>
                <div className="mx-auto bg-gradient-primary rounded-full p-4 w-16 h-16 flex items-center justify-center mb-4">
                  <Heart className="h-8 w-8 text-primary-foreground" />
                </div>
                <CardTitle className="text-primary">Qualidade premium</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Apenas os melhores produtos de marcas confiáveis para seus pets amados
                </CardDescription>
              </CardContent>
            </Card>

            <Card className="text-center shadow-soft hover:shadow-secondary transition-all duration-300">
              <CardHeader>
                <div className="mx-auto bg-gradient-secondary rounded-full p-4 w-16 h-16 flex items-center justify-center mb-4">
                  <Shield className="h-8 w-8 text-secondary-foreground" />
                </div>
                <CardTitle className="text-secondary">Saúde garantida</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Todos os nossos produtos são aprovados por veterinários e vêm com garantias de
                  saúde
                </CardDescription>
              </CardContent>
            </Card>

            <Card className="text-center shadow-soft hover:shadow-accent transition-all duration-300">
              <CardHeader>
                <div className="mx-auto bg-gradient-accent rounded-full p-4 w-16 h-16 flex items-center justify-center mb-4">
                  <Truck className="h-8 w-8 text-accent-foreground" />
                </div>
                <CardTitle className="text-accent-foreground">Entrega rápida</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Frete grátis em pedidos acima de R$ 250. Entrega no mesmo dia disponível
                </CardDescription>
              </CardContent>
            </Card>

            <Card className="text-center shadow-soft hover:shadow-primary transition-all duration-300">
              <CardHeader>
                <div className="mx-auto bg-gradient-primary rounded-full p-4 w-16 h-16 flex items-center justify-center mb-4">
                  <Users className="h-8 w-8 text-primary-foreground" />
                </div>
                <CardTitle className="text-primary">Suporte especializado</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription>
                  Nossos especialistas em cuidados com pets estão aqui para ajudá-lo a fazer as
                  melhores escolhas
                </CardDescription>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      {/* Popular Categories */}
      <section className="py-20">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-foreground mb-4">Compre por pet</h2>
            <p className="text-lg text-muted-foreground">
              Encontre tudo que seu pet específico precisa
            </p>
          </div>

          {loadingCategories ? (
            <div className="flex justify-center items-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
              <span className="ml-2 text-muted-foreground">Carregando categorias...</span>
            </div>
          ) : (
            <div className="grid sm:grid-cols-2 lg:grid-cols-5 gap-6">
              {categories.map((category) => {
                const config = categoryConfig[category.toLowerCase()] || {
                  name: category,
                  icon: null,
                  emoji: "🐾",
                };
                return (
                  <Link key={category} to={`/products/${category}`}>
                    <Card className="group cursor-pointer hover:scale-105 transition-all duration-300 shadow-soft hover:shadow-primary border-2 border-transparent hover:border-primary/20">
                      <CardContent className="p-8 text-center">
                        <div className="mb-4 flex justify-center">
                          {config.icon ? (
                            <div className="bg-gradient-primary rounded-full p-4 text-primary-foreground group-hover:scale-110 transition-transform duration-300">
                              {config.icon}
                            </div>
                          ) : (
                            <span className="text-5xl group-hover:scale-110 transition-transform duration-300">
                              {config.emoji}
                            </span>
                          )}
                        </div>
                        <h3 className="font-semibold text-xl text-foreground group-hover:text-primary transition-colors">
                          {config.name}
                        </h3>
                      </CardContent>
                    </Card>
                  </Link>
                );
              })}
            </div>
          )}
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 bg-gradient-hero">
        <div className="container mx-auto px-4 text-center">
          <div className="max-w-3xl mx-auto">
            <h2 className="text-4xl lg:text-5xl font-bold text-primary-foreground mb-6">
              Pronto para mimar seu pet?
            </h2>
            <p className="text-xl text-primary-foreground/90 mb-8 leading-relaxed">
              Junte-se a milhares de tutores felizes que confiam em nós para a felicidade e saúde de
              seus pets
            </p>
            <Button
              size="lg"
              className="bg-background text-primary hover:bg-background/90 px-8"
              asChild
            >
              <Link to="/products/dogs">Começar a comprar</Link>
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
};

export default Index;
