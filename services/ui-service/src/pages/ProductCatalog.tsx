import { Filter, Grid3X3, Heart, List, Loader2, Search, ShoppingCart, Star } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import Header from "@/components/Header";
import { ProductImageWithFallback } from "@/components/ProductImage";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useCart } from "@/hooks/use-cart";
import { useProductsPaginated } from "@/hooks/use-products";
import { toast } from "@/hooks/use-toast";
import { designSystemStyles } from "@/styles/design-system";

const ProductCatalog = () => {
  const { category } = useParams();
  const [searchQuery, setSearchQuery] = useState("");
  const [sortBy, setSortBy] = useState("popularity");
  const [filterBy, setFilterBy] = useState("all");
  const [viewMode, setViewMode] = useState("grid");
  const [favorites, setFavorites] = useState<number[]>([]);
  const [page, setPage] = useState(1);
  const [isVisible, setIsVisible] = useState(false);
  const pageSize = 12;

  const { products, loading, error, totalCount, totalPages } = useProductsPaginated(
    page,
    pageSize,
    category
  );
  const { addToCart, getItemQuantity } = useCart();

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

    const elements = document.querySelectorAll(".observe-animation");
    elements.forEach((element) => observer.observe(element));

    return () => observer.disconnect();
  }, [products]);

  const toggleFavorite = (productId: string) => {
    const numId = parseInt(productId, 10);
    setFavorites((prev) =>
      prev.includes(numId) ? prev.filter((id) => id !== numId) : [...prev, numId]
    );
  };

  const handleAdicionarToCart = (product: any) => {
    addToCart(product);
    toast({
      title: "Adicionado ao carrinho!",
      description: `${product.name} foi adicionado ao seu carrinho.`,
    });
  };

  const filteredProdutos = products.filter((product) => {
    const matchesSearch =
      product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (product.brand || "").toLowerCase().includes(searchQuery.toLowerCase());
    const matchesFilter =
      filterBy === "all" ||
      (filterBy === "on-sale" && product.price) ||
      (filterBy === "in-stock" && product.quantity > 0);
    return matchesSearch && matchesFilter;
  });

  const categoryNames: Record<string, string> = {
    dogs: "Cães",
    cats: "Gatos",
    birds: "Pássaros",
    fish: "Peixes",
    "small-pets": "Pets pequenos",
    reptiles: "Répteis",
    rabbits: "Coelhos",
  };

  const categoryName = category ? categoryNames[category] || category : "Todos os produtos";

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-[#F8FAF5]">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-12">
          {/* Breadcrumb */}
          <nav className={`mb-8 ${isVisible ? 'page-enter active' : 'page-enter'}`}>
            <div className="flex items-center space-x-2 text-sm font-body text-[#2D6A4F]">
              <Link to="/" className="hover:text-[#52B788] transition-colors">
                Início
              </Link>
              <span>/</span>
              <Link to="/products" className="hover:text-[#52B788] transition-colors">
                Produtos
              </Link>
              {category && (
                <>
                  <span>/</span>
                  <span className="text-[#1B4332] font-semibold">{categoryName}</span>
                </>
              )}
            </div>
          </nav>

          {/* Header */}
          <div className={`mb-12 ${isVisible ? 'hero-enter active' : 'hero-enter'}`}>
            <span className="font-body text-[#52B788] font-semibold text-sm tracking-widest uppercase mb-4 block">
              {category ? `Categoria: ${categoryName}` : "Catálogo Completo"}
            </span>
            <h1 className="font-display text-5xl lg:text-6xl font-bold text-[#1B4332] mb-4">
              {category ? `Produtos para ${categoryName}` : "Todos os produtos"}
            </h1>
            <div className="w-20 h-1 bg-gradient-to-r from-[#52B788] to-[#A7C957] mb-6" />
            <p className="font-body text-xl text-[#2D6A4F]">
              {loading ? (
                <span className="inline-flex items-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Carregando produtos...
                </span>
              ) : (
                <>
                  <span className="font-bold text-[#52B788]">{totalCount || 0}</span> produtos encontrados
                </>
              )}
            </p>
          </div>

          {/* Filters & Search */}
          <Card className="mb-12 shadow-lg border-2 border-[#1B4332]/10 observe-animation rounded-2xl">
            <CardContent className="p-6">
              <div className="flex flex-col lg:flex-row gap-4 items-center justify-between">
                <div className="flex flex-col sm:flex-row gap-4 flex-1 w-full">
                  {/* Search */}
                  <div className="relative flex-1 max-w-md">
                    <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-[#2D6A4F]" />
                    <Input
                      placeholder="Buscar produtos..."
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      className="pl-12 font-body border-2 border-[#1B4332]/10 rounded-xl focus:border-[#52B788] h-12"
                    />
                  </div>

                  {/* Filter */}
                  <Select value={filterBy} onValueChange={setFilterBy}>
                    <SelectTrigger className="w-full sm:w-48 font-body border-2 border-[#1B4332]/10 rounded-xl h-12">
                      <Filter className="h-4 w-4 mr-2 text-[#52B788]" />
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent className="rounded-2xl">
                      <SelectItem value="all">Todos os produtos</SelectItem>
                      <SelectItem value="on-sale">Em promoção</SelectItem>
                      <SelectItem value="in-stock">Em estoque</SelectItem>
                    </SelectContent>
                  </Select>

                  {/* Sort */}
                  <Select value={sortBy} onValueChange={setSortBy}>
                    <SelectTrigger className="w-full sm:w-48 font-body border-2 border-[#1B4332]/10 rounded-xl h-12">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent className="rounded-2xl">
                      <SelectItem value="popularity">Mais Popular</SelectItem>
                      <SelectItem value="price-low">Preço: Menor ao Maior</SelectItem>
                      <SelectItem value="price-high">Preço: Maior ao Menor</SelectItem>
                      <SelectItem value="rating">Melhor Avaliado</SelectItem>
                      <SelectItem value="newest">Mais Recente</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {/* View Mode */}
                <div className="flex items-center space-x-2">
                  <Button
                    variant={viewMode === "grid" ? "default" : "outline"}
                    size="icon"
                    onClick={() => setViewMode("grid")}
                    className={viewMode === "grid" ? "bg-[#52B788] hover:bg-[#40916C]" : "hover:bg-[#52B788]/10"}
                  >
                    <Grid3X3 className="h-4 w-4" />
                  </Button>
                  <Button
                    variant={viewMode === "list" ? "default" : "outline"}
                    size="icon"
                    onClick={() => setViewMode("list")}
                    className={viewMode === "list" ? "bg-[#52B788] hover:bg-[#40916C]" : "hover:bg-[#52B788]/10"}
                  >
                    <List className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Products Grid */}
          <div
            className={
              viewMode === "grid"
                ? "grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6"
                : "space-y-6"
            }
          >
            {loading ? (
              <div className="col-span-full flex flex-col items-center justify-center py-20">
                <Loader2 className="h-16 w-16 animate-spin text-[#52B788] mb-4" />
                <p className="font-body text-[#2D6A4F] text-lg">Carregando produtos...</p>
              </div>
            ) : error ? (
              <div className="col-span-full">
                <Card className="text-center py-16 rounded-2xl border-2 border-[#1B4332]/10">
                  <CardContent>
                    <h3 className="font-display text-2xl font-bold text-[#1B4332] mb-2">
                      Erro ao carregar produtos
                    </h3>
                    <p className="font-body text-[#2D6A4F] mb-6">{error}</p>
                    <Button
                      onClick={() => window.location.reload()}
                      className="btn-primary-custom font-body px-8 py-3 rounded-full"
                    >
                      Tentar novamente
                    </Button>
                  </CardContent>
                </Card>
              </div>
            ) : (
              filteredProdutos.map((product, index) => (
                <Card
                  key={product._id}
                  className="product-card observe-animation bg-white shadow-lg hover:shadow-2xl rounded-3xl overflow-hidden"
                  style={{ animationDelay: `${index * 0.05}s` }}
                >
                  <CardContent className="p-0">
                    {viewMode === "grid" ? (
                      <>
                        <div className="relative group">
                          <div className="aspect-square overflow-hidden">
                            <ProductImageWithFallback
                              images={product.images || []}
                              alt={product.name}
                              className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
                              fallbackIcon="🐕"
                            />
                          </div>
                          <Button
                            variant="ghost"
                            size="icon"
                            className={`absolute top-3 right-3 rounded-full bg-white/90 backdrop-blur-sm shadow-lg ${
                              favorites.includes(parseInt(product._id, 10))
                                ? "text-red-500 hover:text-red-600"
                                : "text-[#2D6A4F] hover:text-red-500"
                            }`}
                            onClick={() => toggleFavorite(product._id)}
                          >
                            <Heart
                              className={`h-5 w-5 ${favorites.includes(parseInt(product._id, 10)) ? "fill-current" : ""}`}
                            />
                          </Button>
                          {product.price > 100 && (
                            <Badge className="absolute top-3 left-3 bg-[#A7C957] text-[#1B4332] font-bold rounded-full px-4 py-1">
                              15% OFF
                            </Badge>
                          )}
                          {product.quantity === 0 && (
                            <div className="absolute inset-0 bg-white/90 backdrop-blur-sm flex items-center justify-center">
                              <Badge className="bg-[#1B4332] text-white px-6 py-2 text-base">Sem Estoque</Badge>
                            </div>
                          )}
                        </div>

                        <div className="p-6 space-y-3">
                          {product.brand && (
                            <p className="font-body text-xs text-[#95D5B2] font-semibold uppercase tracking-widest">
                              {product.brand}
                            </p>
                          )}
                          <Link to={`/product/${product._id}`} className="block">
                            <h3 className="font-display text-lg font-bold text-[#1B4332] hover:text-[#52B788] transition-colors line-clamp-2 min-h-[3.5rem]">
                              {product.name}
                            </h3>
                          </Link>

                          <div className="flex items-center gap-1">
                            {[1, 2, 3, 4, 5].map((star) => (
                              <Star
                                key={star}
                                className={`h-4 w-4 ${
                                  star <= Math.round(product.rating || 0)
                                    ? "text-[#A7C957] fill-[#A7C957]"
                                    : "text-gray-300"
                                }`}
                              />
                            ))}
                            <span className="font-body text-sm font-semibold text-[#1B4332] ml-2">
                              {(product.rating || 0).toFixed(1)}
                            </span>
                          </div>

                          {product.category && (
                            <Badge className="bg-[#95D5B2]/20 text-[#2D6A4F] border-0 font-body">
                              {product.category}
                            </Badge>
                          )}

                          <div className="space-y-3 pt-2">
                            <div className="flex items-baseline gap-2">
                              <span className="font-display text-3xl font-bold text-[#52B788]">
                                R$ {product.price.toFixed(2)}
                              </span>
                              {product.price > 100 && (
                                <span className="font-body text-sm text-[#2D6A4F] line-through">
                                  R$ {(product.price * 1.15).toFixed(2)}
                                </span>
                              )}
                            </div>
                            {product.quantity < 10 && product.quantity > 0 && (
                              <div className="flex items-center gap-2">
                                <div className="w-2 h-2 bg-orange-500 rounded-full animate-pulse" />
                                <span className="font-body text-xs font-semibold text-orange-600">
                                  Apenas {product.quantity} em estoque
                                </span>
                              </div>
                            )}
                            <Button
                              size="lg"
                              onClick={() => handleAdicionarToCart(product)}
                              disabled={product.quantity === 0}
                              className="w-full btn-primary-custom font-body rounded-full font-semibold"
                            >
                              <ShoppingCart className="h-4 w-4 mr-2" />
                              {(() => {
                                const quantity = getItemQuantity(product._id);
                                if (quantity > 0) {
                                  return `No carrinho (${quantity})`;
                                }
                                return "Adicionar";
                              })()}
                            </Button>
                          </div>
                        </div>
                      </>
                    ) : (
                      <div className="flex gap-6 p-6">
                        <div className="w-48 h-48 flex-shrink-0 rounded-2xl overflow-hidden">
                          <ProductImageWithFallback
                            images={product.images || []}
                            alt={product.name}
                            className="w-full h-full object-cover"
                            fallbackIcon="🐕"
                          />
                        </div>

                        <div className="flex-1 space-y-3">
                          <div className="flex items-start justify-between">
                            <div>
                              {product.brand && (
                                <p className="font-body text-xs text-[#95D5B2] font-semibold uppercase tracking-widest mb-2">
                                  {product.brand}
                                </p>
                              )}
                              <Link to={`/product/${product._id}`}>
                                <h3 className="font-display text-2xl font-bold text-[#1B4332] hover:text-[#52B788] transition-colors">
                                  {product.name}
                                </h3>
                              </Link>
                              {product.description && (
                                <p className="font-body text-[#2D6A4F] mt-2 line-clamp-2">
                                  {product.description}
                                </p>
                              )}
                            </div>
                            <Button
                              variant="ghost"
                              size="icon"
                              className={`rounded-full ${
                                favorites.includes(parseInt(product._id, 10))
                                  ? "text-red-500"
                                  : "text-[#2D6A4F]"
                              }`}
                              onClick={() => toggleFavorite(product._id)}
                            >
                              <Heart
                                className={`h-5 w-5 ${favorites.includes(parseInt(product._id, 10)) ? "fill-current" : ""}`}
                              />
                            </Button>
                          </div>

                          <div className="flex items-center gap-4">
                            <div className="flex items-center gap-1">
                              <Star className="h-5 w-5 text-[#A7C957] fill-[#A7C957]" />
                              <span className="font-body text-sm font-semibold">{product.rating}</span>
                            </div>
                            {product.quantity === 0 && (
                              <Badge className="bg-[#1B4332] text-white">Sem estoque</Badge>
                            )}
                            {product.quantity < 10 && product.quantity > 0 && (
                              <Badge className="border-orange-500 text-orange-600">Estoque baixo</Badge>
                            )}
                          </div>

                          <div className="flex items-center justify-between pt-4">
                            <div className="font-display text-3xl font-bold text-[#52B788]">
                              R$ {product.price.toFixed(2)}
                            </div>
                            <Button
                              className="btn-primary-custom font-body rounded-full px-8"
                              onClick={() => handleAdicionarToCart(product)}
                              disabled={product.quantity === 0}
                            >
                              <ShoppingCart className="h-4 w-4 mr-2" />
                              {(() => {
                                const quantity = getItemQuantity(product._id);
                                if (quantity > 0) {
                                  return `No carrinho (${quantity})`;
                                }
                                return "Adicionar ao carrinho";
                              })()}
                            </Button>
                          </div>
                        </div>
                      </div>
                    )}
                  </CardContent>
                </Card>
              ))
            )}
          </div>

          {/* No Results */}
          {filteredProdutos.length === 0 && !loading && (
            <Card className="text-center py-16 rounded-2xl border-2 border-[#1B4332]/10">
              <CardContent>
                <h3 className="font-display text-3xl font-bold text-[#1B4332] mb-3">
                  Nenhum produto encontrado
                </h3>
                <p className="font-body text-[#2D6A4F] text-lg mb-8">
                  Tente ajustar sua busca ou critérios de filtro
                </p>
                <Button
                  onClick={() => {
                    setSearchQuery("");
                    setFilterBy("all");
                  }}
                  className="btn-primary-custom font-body px-8 py-3 rounded-full"
                >
                  Limpar filtros
                </Button>
              </CardContent>
            </Card>
          )}

          {/* Pagination */}
          {totalPages > 1 && filteredProdutos.length > 0 && (
            <div className="mt-16 flex justify-center items-center gap-3">
              <Button
                variant="outline"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="font-body border-2 border-[#1B4332] hover:bg-[#1B4332] hover:text-white rounded-full px-6"
              >
                ← Anterior
              </Button>
              <div className="flex items-center gap-2">
                {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                  const pageNum = i + 1;
                  return (
                    <Button
                      key={pageNum}
                      variant={page === pageNum ? "default" : "outline"}
                      onClick={() => setPage(pageNum)}
                      className={
                        page === pageNum
                          ? "bg-[#52B788] hover:bg-[#40916C] rounded-full"
                          : "border-2 border-[#1B4332]/20 rounded-full hover:border-[#52B788]"
                      }
                    >
                      {pageNum}
                    </Button>
                  );
                })}
                {totalPages > 5 && <span className="font-body text-[#2D6A4F]">...</span>}
              </div>
              <Button
                variant="outline"
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="font-body border-2 border-[#1B4332] hover:bg-[#1B4332] hover:text-white rounded-full px-6"
              >
                Próximo →
              </Button>
            </div>
          )}
        </main>
      </div>
    </>
  );
};

export default ProductCatalog;
