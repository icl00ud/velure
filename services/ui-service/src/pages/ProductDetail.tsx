import {
  ArrowLeft,
  Award,
  Heart,
  Minus,
  Plus,
  RotateCcw,
  Share2,
  Shield,
  ShoppingCart,
  Star,
  Truck,
} from "lucide-react";
import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import Header from "@/components/Header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toast } from "@/hooks/use-toast";

// Mock product data
const mockProduct = {
  id: 1,
  name: "Ração Premium para Cães - Frango e Arroz",
  brand: "PetNutrition Pro",
  price: 45.99,
  originalPrice: 52.99,
  rating: 4.8,
  reviews: 156,
  category: "Ração",
  tags: ["Premium", "Sem Grãos", "Adulto"],
  inStock: true,
  stockQuantity: 23,
  discount: 13,
  description:
    "Ofereça ao seu cão a nutrição que ele merece com nossa fórmula premium de frango e arroz. Feita com frango de verdade como ingrediente principal, esta receita fornece nutrição completa e equilibrada para cães adultos.",
  features: [
    "Frango de verdade como ingrediente principal",
    "Sem corantes, sabores ou conservantes artificiais",
    "Rico em proteínas para manutenção muscular",
    "Vitaminas e minerais adicionados para suporte imunológico",
    "Ácidos graxos ômega-6 para pele e pelo saudáveis",
  ],
  specifications: {
    Peso: "6,8 kg",
    "Fase de Vida": "Adulto",
    "Porte": "Todos os Portes",
    "Proteína Principal": "Frango",
    "Dieta Especial": "Sem Grãos",
  },
  ingredients:
    "Frango Desossado, Farinha de Frango, Batata Doce, Ervilhas, Gordura de Frango, Polpa de Tomate, Sabor Natural, Sal, Cloreto de Colina, Taurina, Raiz de Chicória Seca, Extrato de Yucca Schidigera, Extrato de Alecrim, Tocoferóis Mistos",
  images: [
    "/api/placeholder/500/500",
    "/api/placeholder/500/500",
    "/api/placeholder/500/500",
    "/api/placeholder/500/500",
  ],
};

const mockReviews = [
  {
    id: 1,
    author: "Maria S.",
    rating: 5,
    date: "2024-01-15",
    title: "Meu cachorro adora!",
    content:
      "Meu golden retriever adora essa ração. O pelo dele está mais brilhante e ele tem mais energia. Recomendo muito!",
  },
  {
    id: 2,
    author: "Pedro R.",
    rating: 4,
    date: "2024-01-10",
    title: "Ótima qualidade",
    content:
      "Ração de boa qualidade com preço justo. Meu cachorro levou alguns dias para se adaptar, mas agora come feliz.",
  },
  {
    id: 3,
    author: "Ana K.",
    rating: 5,
    date: "2024-01-05",
    title: "Excelente nutrição",
    content:
      "O veterinário recomendou essa marca. Ótimos ingredientes e a digestão do meu cachorro melhorou significativamente.",
  },
];

const ProductDetail = () => {
  const { id } = useParams();
  const [selectedImage, setSelectedImage] = useState(0);
  const [quantity, setQuantity] = useState(1);
  const [isFavorite, setIsFavorite] = useState(false);

  const handleAddToCart = () => {
    toast({
      title: "Adicionado ao carrinho!",
      description: `${quantity} x ${mockProduct.name} adicionado ao seu carrinho.`,
    });
  };

  const handleBuyNow = () => {
    toast({
      title: "Redirecionando para o checkout...",
      description: "Levando você para o pagamento seguro.",
    });
  };

  return (
    <div className="min-h-screen bg-background">
      <Header />

      <main className="container mx-auto px-4 py-8">
        {/* Breadcrumb */}
        <nav className="mb-6">
          <div className="flex items-center space-x-2 text-sm text-muted-foreground">
            <Link to="/" className="hover:text-primary">
              Início
            </Link>
            <span>/</span>
            <Link to="/products/dogs" className="hover:text-primary">
              Cães
            </Link>
            <span>/</span>
            <span className="text-foreground font-medium">{mockProduct.category}</span>
          </div>
        </nav>

        <Link
          to="/products/dogs"
          className="inline-flex items-center text-muted-foreground hover:text-primary transition-colors mb-6"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Voltar para produtos
        </Link>

        <div className="grid lg:grid-cols-2 gap-12 mb-12">
          {/* Product Images */}
          <div className="space-y-4">
            <div className="aspect-square bg-muted rounded-lg flex items-center justify-center text-8xl overflow-hidden">
              🐕
            </div>

            <div className="grid grid-cols-4 gap-2">
              {[0, 1, 2, 3].map((index) => (
                <button
                  key={index}
                  onClick={() => setSelectedImage(index)}
                  className={`aspect-square bg-muted rounded-lg flex items-center justify-center text-2xl border-2 transition-colors ${
                    selectedImage === index
                      ? "border-primary"
                      : "border-transparent hover:border-muted-foreground"
                  }`}
                >
                  🐕
                </button>
              ))}
            </div>
          </div>

          {/* Product Info */}
          <div className="space-y-6">
            <div>
              <p className="text-muted-foreground font-medium mb-2">{mockProduct.brand}</p>
              <h1 className="text-3xl font-bold text-foreground mb-4">{mockProduct.name}</h1>

              <div className="flex items-center space-x-4 mb-4">
                <div className="flex items-center">
                  {[1, 2, 3, 4, 5].map((star) => (
                    <Star
                      key={star}
                      className={`h-4 w-4 ${
                        star <= Math.floor(mockProduct.rating)
                          ? "text-accent fill-current"
                          : "text-muted-foreground"
                      }`}
                    />
                  ))}
                  <span className="ml-2 font-medium">{mockProduct.rating}</span>
                </div>
                <span className="text-muted-foreground">({mockProduct.reviews} avaliações)</span>
              </div>

              <div className="flex flex-wrap gap-2 mb-6">
                {mockProduct.tags.map((tag, index) => (
                  <Badge key={index} variant="secondary">
                    {tag}
                  </Badge>
                ))}
              </div>
            </div>

            <div className="flex items-baseline space-x-4">
              <span className="text-3xl font-bold text-primary">${mockProduct.price}</span>
              {mockProduct.originalPrice && (
                <span className="text-xl text-muted-foreground line-through">
                  ${mockProduct.originalPrice}
                </span>
              )}
              {mockProduct.discount > 0 && (
                <Badge className="bg-secondary text-secondary-foreground">
                  Economize {mockProduct.discount}%
                </Badge>
              )}
            </div>

            <p className="text-muted-foreground leading-relaxed">{mockProduct.description}</p>

            {/* Stock Status */}
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-primary rounded-full"></div>
              <span className="text-sm font-medium text-primary">
                Em estoque ({mockProduct.stockQuantity} disponíveis)
              </span>
            </div>

            {/* Quantity & Actions */}
            <div className="space-y-4">
              <div className="flex items-center space-x-4">
                <span className="font-medium">Quantidade:</span>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setQuantity(Math.max(1, quantity - 1))}
                    disabled={quantity <= 1}
                  >
                    <Minus className="h-4 w-4" />
                  </Button>
                  <span className="w-12 text-center font-medium">{quantity}</span>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setQuantity(Math.min(mockProduct.stockQuantity, quantity + 1))}
                    disabled={quantity >= mockProduct.stockQuantity}
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                </div>
              </div>

              <div className="flex space-x-4">
                <Button
                  onClick={handleAddToCart}
                  className="flex-1 bg-gradient-primary hover:opacity-90 text-primary-foreground"
                >
                  <ShoppingCart className="h-4 w-4 mr-2" />
                  Adicionar ao carrinho
                </Button>

                <Button
                  variant="outline"
                  onClick={handleBuyNow}
                  className="flex-1 border-primary text-primary hover:bg-primary hover:text-primary-foreground"
                >
                  Comprar agora
                </Button>
              </div>

              <div className="flex space-x-2">
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => setIsFavorite(!isFavorite)}
                  className={isFavorite ? "text-red-500 border-red-500" : ""}
                >
                  <Heart className={`h-4 w-4 ${isFavorite ? "fill-current" : ""}`} />
                </Button>
                <Button variant="outline" size="icon">
                  <Share2 className="h-4 w-4" />
                </Button>
              </div>
            </div>

            {/* Benefits */}
            <div className="grid grid-cols-2 gap-4 pt-6">
              <div className="flex items-center space-x-3">
                <div className="bg-primary/10 rounded-full p-2">
                  <Truck className="h-4 w-4 text-primary" />
                </div>
                <div>
                  <p className="text-sm font-medium">Frete grátis</p>
                  <p className="text-xs text-muted-foreground">Em pedidos acima de R$ 100</p>
                </div>
              </div>

              <div className="flex items-center space-x-3">
                <div className="bg-secondary/10 rounded-full p-2">
                  <RotateCcw className="h-4 w-4 text-secondary" />
                </div>
                <div>
                  <p className="text-sm font-medium">Devolução em 30 dias</p>
                  <p className="text-xs text-muted-foreground">Garantia de dinheiro de volta</p>
                </div>
              </div>

              <div className="flex items-center space-x-3">
                <div className="bg-accent/10 rounded-full p-2">
                  <Shield className="h-4 w-4 text-accent-foreground" />
                </div>
                <div>
                  <p className="text-sm font-medium">Garantia de qualidade</p>
                  <p className="text-xs text-muted-foreground">Produtos premium</p>
                </div>
              </div>

              <div className="flex items-center space-x-3">
                <div className="bg-primary/10 rounded-full p-2">
                  <Award className="h-4 w-4 text-primary" />
                </div>
                <div>
                  <p className="text-sm font-medium">Aprovado por veterinários</p>
                  <p className="text-xs text-muted-foreground">Confiável por veterinários</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Product Details Tabs */}
        <Card className="shadow-soft">
          <CardContent className="p-0">
            <Tabs defaultValue="description" className="w-full">
              <TabsList className="grid w-full grid-cols-4">
                <TabsTrigger value="description">Descrição</TabsTrigger>
                <TabsTrigger value="specifications">Especificações</TabsTrigger>
                <TabsTrigger value="ingredients">Ingredientes</TabsTrigger>
                <TabsTrigger value="reviews">Avaliações ({mockProduct.reviews})</TabsTrigger>
              </TabsList>

              <TabsContent value="description" className="p-6">
                <div className="space-y-4">
                  <h3 className="text-xl font-semibold text-foreground">Descrição do Produto</h3>
                  <p className="text-muted-foreground leading-relaxed">{mockProduct.description}</p>

                  <h4 className="text-lg font-semibold text-foreground mt-6">Principais Características</h4>
                  <ul className="space-y-2">
                    {mockProduct.features.map((feature, index) => (
                      <li key={index} className="flex items-start space-x-2">
                        <div className="w-1.5 h-1.5 bg-primary rounded-full mt-2 flex-shrink-0"></div>
                        <span className="text-muted-foreground">{feature}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </TabsContent>

              <TabsContent value="specifications" className="p-6">
                <div className="space-y-4">
                  <h3 className="text-xl font-semibold text-foreground">Especificações</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {Object.entries(mockProduct.specifications).map(([key, value]) => (
                      <div key={key} className="flex justify-between py-2 border-b border-border">
                        <span className="font-medium text-foreground">{key}:</span>
                        <span className="text-muted-foreground">{value}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="ingredients" className="p-6">
                <div className="space-y-4">
                  <h3 className="text-xl font-semibold text-foreground">Ingredientes</h3>
                  <p className="text-muted-foreground leading-relaxed">{mockProduct.ingredients}</p>
                </div>
              </TabsContent>

              <TabsContent value="reviews" className="p-6">
                <div className="space-y-6">
                  <div className="flex items-center justify-between">
                    <h3 className="text-xl font-semibold text-foreground">Avaliações de Clientes</h3>
                    <div className="flex items-center space-x-2">
                      <div className="flex items-center">
                        {[1, 2, 3, 4, 5].map((star) => (
                          <Star
                            key={star}
                            className={`h-4 w-4 ${
                              star <= Math.floor(mockProduct.rating)
                                ? "text-accent fill-current"
                                : "text-muted-foreground"
                            }`}
                          />
                        ))}
                      </div>
                      <span className="font-medium">{mockProduct.rating} de 5</span>
                    </div>
                  </div>

                  <Separator />

                  <div className="space-y-6">
                    {mockReviews.map((review) => (
                      <div key={review.id} className="space-y-3">
                        <div className="flex items-start justify-between">
                          <div>
                            <div className="flex items-center space-x-2 mb-1">
                              <span className="font-medium text-foreground">{review.author}</span>
                              <div className="flex items-center">
                                {[1, 2, 3, 4, 5].map((star) => (
                                  <Star
                                    key={star}
                                    className={`h-3 w-3 ${
                                      star <= review.rating
                                        ? "text-accent fill-current"
                                        : "text-muted-foreground"
                                    }`}
                                  />
                                ))}
                              </div>
                            </div>
                            <h4 className="font-medium text-foreground">{review.title}</h4>
                          </div>
                          <span className="text-sm text-muted-foreground">{review.date}</span>
                        </div>
                        <p className="text-muted-foreground">{review.content}</p>
                        <Separator />
                      </div>
                    ))}
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </main>
    </div>
  );
};

export default ProductDetail;
