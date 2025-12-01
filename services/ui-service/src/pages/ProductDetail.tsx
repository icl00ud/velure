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
import { useState } from "react";
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

const ProductDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [selectedImage, setSelectedImage] = useState(0);
  const [quantity, setQuantity] = useState(1);
  const [isFavorite, setIsFavorite] = useState(false);

  const { product, loading, error } = useProduct(id || "");
  const { addToCart, getItemQuantity } = useCart();

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
      <div className="min-h-screen bg-background">
        <Header />
        <div className="flex items-center justify-center py-24">
          <Loader2 className="h-12 w-12 animate-spin text-primary" />
          <span className="ml-4 text-lg text-muted-foreground">Carregando produto...</span>
        </div>
      </div>
    );
  }

  if (error || !product) {
    return (
      <div className="min-h-screen bg-background">
        <Header />
        <main className="container mx-auto px-4 py-8">
          <Card className="text-center py-12">
            <CardContent>
              <h3 className="text-xl font-semibold text-foreground mb-2">Produto não encontrado</h3>
              <p className="text-muted-foreground mb-6">
                {error || "O produto que você está procurando não existe."}
              </p>
              <Button
                asChild
                className="bg-gradient-primary hover:opacity-90 text-primary-foreground"
              >
                <Link to="/products">Ver todos os produtos</Link>
              </Button>
            </CardContent>
          </Card>
        </main>
      </div>
    );
  }

  const inStock = product.quantity > 0;
  const cartQuantity = getItemQuantity(product._id);

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
            <Link to="/products" className="hover:text-primary">
              Produtos
            </Link>
            {product.category && (
              <>
                <span>/</span>
                <Link to={`/products/${product.category}`} className="hover:text-primary">
                  {product.category}
                </Link>
              </>
            )}
            <span>/</span>
            <span className="text-foreground font-medium line-clamp-1">{product.name}</span>
          </div>
        </nav>

        <Link
          to="/products"
          className="inline-flex items-center text-muted-foreground hover:text-primary transition-colors mb-6"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Voltar para produtos
        </Link>

        <div className="grid lg:grid-cols-2 gap-12 mb-12">
          {/* Product Images */}
          <div className="space-y-4">
            <div className="aspect-square bg-muted rounded-lg overflow-hidden">
              <ProductImageWithFallback
                images={product.images || []}
                alt={product.name}
                className="w-full h-full object-cover"
                fallbackIcon="🐕"
              />
            </div>

            {product.images && product.images.length > 1 && (
              <div className="grid grid-cols-4 gap-2">
                {product.images.slice(0, 4).map((image, index) => (
                  <button
                    type="button"
                    key={`image-${image}-${index}`}
                    onClick={() => setSelectedImage(index)}
                    className={`aspect-square bg-muted rounded-lg overflow-hidden border-2 transition-colors ${
                      selectedImage === index
                        ? "border-primary"
                        : "border-transparent hover:border-muted-foreground"
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
          <div className="space-y-6">
            <div>
              {product.brand && (
                <p className="text-muted-foreground font-medium mb-2">{product.brand}</p>
              )}
              <h1 className="text-3xl font-bold text-foreground mb-4">{product.name}</h1>

              <div className="flex items-center space-x-4 mb-4">
                <div className="flex items-center">
                  {[1, 2, 3, 4, 5].map((star) => (
                    <Star
                      key={star}
                      className={`h-4 w-4 ${
                        star <= Math.floor(product.rating || 0)
                          ? "text-yellow-400 fill-yellow-400"
                          : "text-muted-foreground"
                      }`}
                    />
                  ))}
                  <span className="ml-2 font-medium">{(product.rating || 0).toFixed(1)}</span>
                </div>
              </div>

              <div className="flex flex-wrap gap-2 mb-6">
                {product.category && <Badge variant="secondary">{product.category}</Badge>}
                {product.colors?.map((color) => (
                  <Badge key={`color-${color}`} variant="outline">
                    {color}
                  </Badge>
                ))}
              </div>
            </div>

            <div className="flex items-baseline space-x-4">
              <span className="text-3xl font-bold text-primary">R$ {product.price.toFixed(2)}</span>
            </div>

            {product.description && (
              <p className="text-muted-foreground leading-relaxed">{product.description}</p>
            )}

            {/* Stock Status */}
            <div className="flex items-center space-x-2">
              <div
                className={`w-2 h-2 rounded-full ${inStock ? "bg-green-500" : "bg-red-500"}`}
              ></div>
              <span
                className={`text-sm font-medium ${inStock ? "text-green-600" : "text-red-600"}`}
              >
                {inStock ? `Em estoque (${product.quantity} disponíveis)` : "Produto esgotado"}
              </span>
            </div>

            {cartQuantity > 0 && (
              <div className="bg-primary/10 rounded-lg p-3 flex items-center justify-between">
                <span className="text-sm text-primary font-medium">
                  Você já tem {cartQuantity} unidade(s) no carrinho
                </span>
                <Button asChild variant="link" className="text-primary p-0 h-auto">
                  <Link to="/cart">Ver carrinho</Link>
                </Button>
              </div>
            )}

            {/* Quantity & Actions */}
            <div className="space-y-4">
              <div className="flex items-center space-x-4">
                <span className="font-medium">Quantidade:</span>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setQuantity(Math.max(1, quantity - 1))}
                    disabled={quantity <= 1 || !inStock}
                  >
                    <Minus className="h-4 w-4" />
                  </Button>
                  <span className="w-12 text-center font-medium">{quantity}</span>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setQuantity(Math.min(product.quantity, quantity + 1))}
                    disabled={quantity >= product.quantity || !inStock}
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                </div>
              </div>

              <div className="flex space-x-4">
                <Button
                  onClick={handleAddToCart}
                  className="flex-1 bg-gradient-primary hover:opacity-90 text-primary-foreground"
                  disabled={!inStock}
                >
                  <ShoppingCart className="h-4 w-4 mr-2" />
                  Adicionar ao carrinho
                </Button>

                <Button
                  variant="outline"
                  onClick={handleBuyNow}
                  className="flex-1 border-primary text-primary hover:bg-primary hover:text-primary-foreground"
                  disabled={!inStock}
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
                >
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
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="description">Descrição</TabsTrigger>
                <TabsTrigger value="specifications">Especificações</TabsTrigger>
              </TabsList>

              <TabsContent value="description" className="p-6">
                <div className="space-y-4">
                  <h3 className="text-xl font-semibold text-foreground">Descrição do Produto</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    {product.description || "Nenhuma descrição disponível para este produto."}
                  </p>
                </div>
              </TabsContent>

              <TabsContent value="specifications" className="p-6">
                <div className="space-y-4">
                  <h3 className="text-xl font-semibold text-foreground">Especificações</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {product.sku && (
                      <div className="flex justify-between py-2 border-b border-border">
                        <span className="font-medium text-foreground">SKU:</span>
                        <span className="text-muted-foreground">{product.sku}</span>
                      </div>
                    )}
                    {product.brand && (
                      <div className="flex justify-between py-2 border-b border-border">
                        <span className="font-medium text-foreground">Marca:</span>
                        <span className="text-muted-foreground">{product.brand}</span>
                      </div>
                    )}
                    {product.category && (
                      <div className="flex justify-between py-2 border-b border-border">
                        <span className="font-medium text-foreground">Categoria:</span>
                        <span className="text-muted-foreground">{product.category}</span>
                      </div>
                    )}
                    {product.dimensions?.weight && (
                      <div className="flex justify-between py-2 border-b border-border">
                        <span className="font-medium text-foreground">Peso:</span>
                        <span className="text-muted-foreground">
                          {product.dimensions.weight} kg
                        </span>
                      </div>
                    )}
                    {product.dimensions?.height && (
                      <div className="flex justify-between py-2 border-b border-border">
                        <span className="font-medium text-foreground">Altura:</span>
                        <span className="text-muted-foreground">
                          {product.dimensions.height} cm
                        </span>
                      </div>
                    )}
                    {product.dimensions?.width && (
                      <div className="flex justify-between py-2 border-b border-border">
                        <span className="font-medium text-foreground">Largura:</span>
                        <span className="text-muted-foreground">{product.dimensions.width} cm</span>
                      </div>
                    )}
                    {product.dimensions?.length && (
                      <div className="flex justify-between py-2 border-b border-border">
                        <span className="font-medium text-foreground">Comprimento:</span>
                        <span className="text-muted-foreground">
                          {product.dimensions.length} cm
                        </span>
                      </div>
                    )}
                    {product.colors && product.colors.length > 0 && (
                      <div className="flex justify-between py-2 border-b border-border">
                        <span className="font-medium text-foreground">Cores disponíveis:</span>
                        <span className="text-muted-foreground">{product.colors.join(", ")}</span>
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
  );
};

export default ProductDetail;
