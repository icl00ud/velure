import {
  ArrowLeft,
  Award,
  Heart,
  Loader2,
  Minus,
  Plus,
  RotateCcw,
  Share2,
  Shield,
  ShoppingCart,
  Star,
  Truck,
} from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import Header from "@/components/Header";
import { ProductImageWithFallback } from "@/components/ProductImage";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useCart } from "@/hooks/use-cart";
import { useProduct } from "@/hooks/use-products";
import { toast } from "@/hooks/use-toast";
import { designSystemStyles } from "@/styles/design-system";

const ProductDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [selectedImage, setSelectedImage] = useState(0);
  const [quantity, setQuantity] = useState(1);
  const [isFavorite, setIsFavorite] = useState(false);
  const [isVisible, setIsVisible] = useState(false);

  const { product, loading, error } = useProduct(id || "");
  const { addToCart, getItemQuantity } = useCart();

  useEffect(() => {
    setIsVisible(true);
  }, []);

  const handleAddToCart = () => {
    if (!product) return;
    addToCart(product, quantity);
    toast({
      title: "Adicionado ao carrinho!",
      description: `${quantity} x ${product.name} adicionado ao seu carrinho.`,
    });
  };

  const handleBuyNow = () => {
    if (!product) return;
    addToCart(product, quantity);
    navigate("/cart");
  };

  if (loading) {
    return (
      <>
        <style>{designSystemStyles}</style>
        <div className="min-h-screen bg-white">
          <Header />
          <div className="flex flex-col items-center justify-center py-32">
            <Loader2 className="h-16 w-16 animate-spin text-[#52B788] mb-4" />
            <span className="font-body text-lg text-[#2D6A4F]">Carregando produto...</span>
          </div>
        </div>
      </>
    );
  }

  if (error || !product) {
    return (
      <>
        <style>{designSystemStyles}</style>
        <div className="min-h-screen bg-white">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-12">
            <Card className="text-center py-20 rounded-3xl border border-slate-200 shadow-2xl">
              <CardContent>
                <h3 className="font-display text-3xl font-bold text-[#1B4332] mb-4">
                  Produto não encontrado
                </h3>
                <p className="font-body text-lg text-[#2D6A4F] mb-8">
                  {error || "O produto que você está procurando não existe."}
                </p>
                <Button
                  asChild
                  className="btn-primary-custom font-body px-10 py-4 rounded-full text-lg"
                >
                  <Link to="/products">Ver todos os produtos</Link>
                </Button>
              </CardContent>
            </Card>
          </main>
        </div>
      </>
    );
  }

  const inStock = product.quantity > 0;
  const cartQuantity = getItemQuantity(product._id);

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-white">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-12">
          {/* Breadcrumb */}
          <nav className={`mb-8 ${isVisible ? "page-enter active" : "page-enter"}`}>
            <div className="flex items-center space-x-2 text-sm font-body text-[#2D6A4F]">
              <Link to="/" className="hover:text-[#52B788] transition-colors">
                Início
              </Link>
              <span>/</span>
              <Link to="/products" className="hover:text-[#52B788] transition-colors">
                Produtos
              </Link>
              {product.category && (
                <>
                  <span>/</span>
                  <Link
                    to={`/products/${product.category}`}
                    className="hover:text-[#52B788] transition-colors"
                  >
                    {product.category}
                  </Link>
                </>
              )}
              <span>/</span>
              <span className="text-[#1B4332] font-semibold line-clamp-1">{product.name}</span>
            </div>
          </nav>

          <Link
            to="/products"
            className="inline-flex items-center font-body text-[#2D6A4F] hover:text-[#52B788] transition-colors mb-8 group"
          >
            <ArrowLeft className="h-5 w-5 mr-2 group-hover:-translate-x-1 transition-transform" />
            Voltar para produtos
          </Link>

          <div className="grid lg:grid-cols-2 gap-12 mb-16">
            {/* Product Images */}
            <div className={`space-y-6 ${isVisible ? "hero-enter active" : "hero-enter"}`}>
              <div className="aspect-square bg-white rounded-3xl overflow-hidden shadow-2xl border border-slate-200 relative group">
                <ProductImageWithFallback
                  images={product.images || []}
                  alt={product.name}
                  className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                  fallbackIcon="🐕"
                />
                {!inStock && (
                  <div className="absolute inset-0 bg-white/90 backdrop-blur-sm flex items-center justify-center">
                    <Badge className="bg-[#1B4332] text-white px-8 py-3 text-lg">
                      Produto Esgotado
                    </Badge>
                  </div>
                )}
              </div>

              {product.images && product.images.length > 1 && (
                <div className="grid grid-cols-4 gap-4">
                  {product.images.slice(0, 4).map((image, index) => (
                    <button
                      type="button"
                      key={`image-${image}-${index}`}
                      onClick={() => setSelectedImage(index)}
                      className={`aspect-square bg-white rounded-2xl overflow-hidden border transition-all duration-300 hover:scale-105 ${
                        selectedImage === index
                          ? "border-[#52B788] shadow-lg"
                          : "border-slate-200 hover:border-[#52B788]/50"
                      }`}
                    >
                      <ProductImageWithFallback
                        images={[image]}
                        alt={`${product.name} - ${index + 1}`}
                        className="w-full h-full object-cover"
                        fallbackIcon="🐕"
                      />
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* Product Info */}
            <div
              className={`space-y-6 ${isVisible ? "hero-enter active" : "hero-enter"}`}
              style={{ animationDelay: "0.2s" }}
            >
              <div>
                {product.brand && (
                  <p className="font-body text-[#95D5B2] font-semibold uppercase tracking-widest text-sm mb-3">
                    {product.brand}
                  </p>
                )}
                <h1 className="font-display text-4xl lg:text-5xl font-bold text-[#1B4332] mb-4">
                  {product.name}
                </h1>

                <div className="flex items-center space-x-4 mb-6">
                  <div className="flex items-center">
                    {[1, 2, 3, 4, 5].map((star) => (
                      <Star
                        key={star}
                        className={`h-5 w-5 ${
                          star <= Math.floor(product.rating || 0)
                            ? "text-[#A7C957] fill-[#A7C957]"
                            : "text-gray-300"
                        }`}
                      />
                    ))}
                    <span className="ml-3 font-body font-bold text-[#1B4332] text-lg">
                      {(product.rating || 0).toFixed(1)}
                    </span>
                  </div>
                </div>

                <div className="flex flex-wrap gap-2 mb-6">
                  {product.category && (
                    <Badge className="bg-[#95D5B2]/20 text-[#2D6A4F] border-0 font-body px-4 py-1">
                      {product.category}
                    </Badge>
                  )}
                  {product.colors?.map((color) => (
                    <Badge
                      key={`color-${color}`}
                      className="border border-slate-200 text-[#1B4332] bg-white font-body px-4 py-1"
                    >
                      {color}
                    </Badge>
                  ))}
                </div>
              </div>

              <div className="flex items-baseline space-x-4 py-6 border-y border-slate-200">
                <span className="font-display text-5xl font-bold text-[#52B788]">
                  R$ {product.price.toFixed(2)}
                </span>
              </div>

              {product.description && (
                <p className="font-body text-lg text-[#2D6A4F] leading-relaxed">
                  {product.description}
                </p>
              )}

              {/* Stock Status */}
              <div className="flex items-center space-x-3 bg-white rounded-2xl p-4 border border-slate-200">
                <div
                  className={`w-3 h-3 rounded-full ${inStock ? "bg-green-500" : "bg-red-500"} animate-pulse`}
                />
                <span
                  className={`font-body font-semibold ${inStock ? "text-green-600" : "text-red-600"}`}
                >
                  {inStock ? `Em estoque (${product.quantity} disponíveis)` : "Produto esgotado"}
                </span>
              </div>

              {cartQuantity > 0 && (
                <div className="bg-gradient-to-r from-[#52B788]/10 to-[#95D5B2]/10 rounded-2xl p-5 flex items-center justify-between border-2 border-[#52B788]/20">
                  <span className="font-body font-bold text-[#52B788]">
                    Você já tem {cartQuantity} unidade(s) no carrinho
                  </span>
                  <Button asChild variant="link" className="text-[#52B788] font-bold p-0 h-auto">
                    <Link to="/cart">Ver carrinho →</Link>
                  </Button>
                </div>
              )}

              {/* Quantity & Actions */}
              <div className="space-y-5">
                <div className="flex items-center space-x-4">
                  <span className="font-body font-bold text-[#1B4332]">Quantidade:</span>
                  <div className="flex items-center space-x-3 bg-slate-100 rounded-full p-1 border border-slate-200">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setQuantity(Math.max(1, quantity - 1))}
                      disabled={quantity <= 1 || !inStock}
                      className="h-10 w-10 rounded-full hover:bg-[#52B788] hover:text-white"
                    >
                      <Minus className="h-4 w-4" />
                    </Button>
                    <span className="w-12 text-center font-body font-bold text-[#1B4332] text-lg">
                      {quantity}
                    </span>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setQuantity(Math.min(product.quantity, quantity + 1))}
                      disabled={quantity >= product.quantity || !inStock}
                      className="h-10 w-10 rounded-full hover:bg-[#52B788] hover:text-white"
                    >
                      <Plus className="h-4 w-4" />
                    </Button>
                  </div>
                </div>

                <div className="flex flex-col sm:flex-row gap-4">
                  <Button
                    onClick={handleAddToCart}
                    className="flex-1 btn-primary-custom font-body text-lg font-semibold rounded-full h-14"
                    disabled={!inStock}
                  >
                    <ShoppingCart className="h-5 w-5 mr-2" />
                    Adicionar ao carrinho
                  </Button>

                  <Button
                    variant="outline"
                    onClick={handleBuyNow}
                    className="flex-1 font-body text-lg font-semibold rounded-full h-14 border border-[#1B4332] hover:bg-[#1B4332] hover:text-white"
                    disabled={!inStock}
                  >
                    Comprar agora
                  </Button>
                </div>

                <div className="flex gap-3">
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setIsFavorite(!isFavorite)}
                    className={`rounded-full h-12 w-12 border ${
                      isFavorite
                        ? "text-red-500 border-red-500 bg-red-50"
                        : "border-slate-200 hover:border-red-500 hover:text-red-500"
                    }`}
                  >
                    <Heart className={`h-5 w-5 ${isFavorite ? "fill-current" : ""}`} />
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => {
                      navigator.clipboard.writeText(window.location.href);
                      toast({
                        title: "Link copiado!",
                        description: "O link do produto foi copiado para a área de transferência.",
                      });
                    }}
                    className="rounded-full h-12 w-12 border border-slate-200 hover:border-[#52B788] hover:text-[#52B788]"
                  >
                    <Share2 className="h-5 w-5" />
                  </Button>
                </div>
              </div>

              {/* Benefits */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-6">
                {[
                  {
                    icon: Truck,
                    title: "Frete grátis",
                    description: "Em pedidos acima de R$ 100",
                    color: "from-[#52B788] to-[#40916C]",
                  },
                  {
                    icon: RotateCcw,
                    title: "Devolução em 30 dias",
                    description: "Garantia de dinheiro de volta",
                    color: "from-[#95D5B2] to-[#2D6A4F]",
                  },
                  {
                    icon: Shield,
                    title: "Garantia de qualidade",
                    description: "Produtos premium",
                    color: "from-[#A7C957] to-[#E5B520]",
                  },
                  {
                    icon: Award,
                    title: "Aprovado por veterinários",
                    description: "Confiável por profissionais",
                    color: "from-[#52B788] to-[#40916C]",
                  },
                ].map((benefit, index) => {
                  const Icon = benefit.icon;
                  return (
                    <div
                      key={index}
                      className="flex items-start space-x-4 bg-white rounded-2xl p-4 border border-slate-200"
                    >
                      <div
                        className={`bg-gradient-to-br ${benefit.color} rounded-xl p-2.5 flex-shrink-0`}
                      >
                        <Icon className="h-5 w-5 text-white" />
                      </div>
                      <div>
                        <p className="font-body font-semibold text-[#1B4332] text-sm">
                          {benefit.title}
                        </p>
                        <p className="font-body text-xs text-[#2D6A4F]">{benefit.description}</p>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>

          {/* Product Details Tabs */}
          <Card className="shadow-2xl border border-slate-200 rounded-3xl observe-animation">
            <CardContent className="p-0">
              <Tabs defaultValue="description" className="w-full">
                <TabsList className="grid w-full grid-cols-2 bg-slate-100 rounded-t-3xl p-2">
                  <TabsTrigger
                    value="description"
                    className="font-body font-semibold rounded-2xl data-[state=active]:bg-white data-[state=active]:text-[#52B788]"
                  >
                    Descrição
                  </TabsTrigger>
                  <TabsTrigger
                    value="specifications"
                    className="font-body font-semibold rounded-2xl data-[state=active]:bg-white data-[state=active]:text-[#52B788]"
                  >
                    Especificações
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="description" className="p-8">
                  <div className="space-y-4">
                    <h3 className="font-display text-3xl font-bold text-[#1B4332]">
                      Descrição do Produto
                    </h3>
                    <div className="w-16 h-1 bg-[#52B788]" />
                    <p className="font-body text-lg text-[#2D6A4F] leading-relaxed">
                      {product.description || "Nenhuma descrição disponível para este produto."}
                    </p>
                  </div>
                </TabsContent>

                <TabsContent value="specifications" className="p-8">
                  <div className="space-y-6">
                    <div>
                      <h3 className="font-display text-3xl font-bold text-[#1B4332]">
                        Especificações
                      </h3>
                      <div className="w-16 h-1 bg-[#52B788] mt-4" />
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {product.sku && (
                        <div className="flex justify-between py-4 border-b border-slate-200">
                          <span className="font-body font-bold text-[#1B4332]">SKU:</span>
                          <span className="font-body text-[#2D6A4F]">{product.sku}</span>
                        </div>
                      )}
                      {product.brand && (
                        <div className="flex justify-between py-4 border-b border-slate-200">
                          <span className="font-body font-bold text-[#1B4332]">Marca:</span>
                          <span className="font-body text-[#2D6A4F]">{product.brand}</span>
                        </div>
                      )}
                      {product.category && (
                        <div className="flex justify-between py-4 border-b border-slate-200">
                          <span className="font-body font-bold text-[#1B4332]">Categoria:</span>
                          <span className="font-body text-[#2D6A4F]">{product.category}</span>
                        </div>
                      )}
                      {product.dimensions?.weight && (
                        <div className="flex justify-between py-4 border-b border-slate-200">
                          <span className="font-body font-bold text-[#1B4332]">Peso:</span>
                          <span className="font-body text-[#2D6A4F]">
                            {product.dimensions.weight} kg
                          </span>
                        </div>
                      )}
                      {product.dimensions?.height && (
                        <div className="flex justify-between py-4 border-b border-slate-200">
                          <span className="font-body font-bold text-[#1B4332]">Altura:</span>
                          <span className="font-body text-[#2D6A4F]">
                            {product.dimensions.height} cm
                          </span>
                        </div>
                      )}
                      {product.dimensions?.width && (
                        <div className="flex justify-between py-4 border-b border-slate-200">
                          <span className="font-body font-bold text-[#1B4332]">Largura:</span>
                          <span className="font-body text-[#2D6A4F]">
                            {product.dimensions.width} cm
                          </span>
                        </div>
                      )}
                      {product.dimensions?.length && (
                        <div className="flex justify-between py-4 border-b border-slate-200">
                          <span className="font-body font-bold text-[#1B4332]">Comprimento:</span>
                          <span className="font-body text-[#2D6A4F]">
                            {product.dimensions.length} cm
                          </span>
                        </div>
                      )}
                      {product.colors && product.colors.length > 0 && (
                        <div className="flex justify-between py-4 border-b border-slate-200">
                          <span className="font-body font-bold text-[#1B4332]">
                            Cores disponíveis:
                          </span>
                          <span className="font-body text-[#2D6A4F]">
                            {product.colors.join(", ")}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>
        </main>
      </div>
    </>
  );
};

export default ProductDetail;
