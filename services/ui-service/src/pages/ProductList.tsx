import { Heart, Loader2, Search, ShoppingCart, Star } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
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
import { useCategories, useProductsPaginated } from "@/hooks/use-products";
import { toast } from "@/hooks/use-toast";
import { designSystemStyles } from "@/styles/design-system";

const ProductList = () => {
  const [searchQuery, setSearchQuery] = useState("");
  const [sortBy, setSortBy] = useState("popularity");
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [favorites, setFavorites] = useState<number[]>([]);
  const [page, setPage] = useState(1);
  const [isVisible, setIsVisible] = useState(false);
  const pageSize = 12;

  const { categories, loading: loadingCategories } = useCategories();
  const { products, loading, error, totalCount, totalPages } = useProductsPaginated(
    page,
    pageSize,
    selectedCategory !== "all" ? selectedCategory : undefined
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
    const numId = parseInt(productId);
    setFavorites((prev) =>
      prev.includes(numId) ? prev.filter((id) => id !== numId) : [...prev, numId]
    );
  };

  const handleAddToCart = (product: any) => {
    addToCart(product, 1);
    toast({
      title: "Adicionado ao carrinho!",
      description: `${product.name} foi adicionado ao seu carrinho.`,
    });
  };

  const filteredProducts = (products || [])
    .filter(
      (product) =>
        product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (product.brand || "").toLowerCase().includes(searchQuery.toLowerCase())
    )
    .sort((a, b) => {
      switch (sortBy) {
        case "price-low":
          return a.price - b.price;
        case "price-high":
          return b.price - a.price;
        case "name":
          return a.name.localeCompare(b.name);
        case "popularity":
        default:
          return 0;
      }
    });

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-[#FAF7F2]">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-12">
          {/* Breadcrumb */}
          <nav className={`mb-8 ${isVisible ? 'page-enter active' : 'page-enter'}`}>
            <div className="flex items-center space-x-2 text-sm font-body text-[#5A6751]">
              <Link to="/" className="hover:text-[#D97757] transition-colors">
                Início
              </Link>
              <span>/</span>
              <span className="text-[#2D3319] font-semibold">Todos os produtos</span>
            </div>
          </nav>

          {/* Header Section */}
          <div className={`mb-12 ${isVisible ? 'hero-enter active' : 'hero-enter'}`}>
            <span className="font-body text-[#D97757] font-semibold text-sm tracking-widest uppercase mb-4 block">
              Catálogo Completo
            </span>
            <h1 className="font-display text-5xl lg:text-6xl font-bold text-[#2D3319] mb-4">
              Todos os produtos
            </h1>
            <div className="w-20 h-1 bg-gradient-to-r from-[#D97757] to-[#F4C430] mb-6" />
            <p className="font-body text-xl text-[#5A6751]">
              {loading ? (
                <span className="inline-flex items-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Carregando produtos...
                </span>
              ) : (
                <>
                  <span className="font-bold text-[#D97757]">{totalCount || 0}</span> produtos
                  disponíveis para seu pet
                </>
              )}
            </p>
          </div>

          {/* Filters */}
          <Card className="mb-12 shadow-lg border-2 border-[#2D3319]/10 rounded-2xl observe-animation">
            <CardContent className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {/* Search */}
                <div className="relative">
                  <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-[#5A6751]" />
                  <Input
                    placeholder="Buscar produtos..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-12 font-body border-2 border-[#2D3319]/10 rounded-xl h-12 focus:border-[#D97757]"
                  />
                </div>

                {/* Category Filter */}
                <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                  <SelectTrigger className="font-body border-2 border-[#2D3319]/10 rounded-xl h-12">
                    <SelectValue placeholder="Todas as Categorias" />
                  </SelectTrigger>
                  <SelectContent className="rounded-2xl">
                    <SelectItem value="all">Todas as Categorias</SelectItem>
                    {loadingCategories ? (
                      <SelectItem value="loading" disabled>
                        Carregando...
                      </SelectItem>
                    ) : (
                      categories.map((category) => (
                        <SelectItem key={category} value={category}>
                          {category}
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </Select>

                {/* Sort By */}
                <Select value={sortBy} onValueChange={setSortBy}>
                  <SelectTrigger className="font-body border-2 border-[#2D3319]/10 rounded-xl h-12">
                    <SelectValue placeholder="Ordenar por" />
                  </SelectTrigger>
                  <SelectContent className="rounded-2xl">
                    <SelectItem value="popularity">Popularidade</SelectItem>
                    <SelectItem value="price-low">Preço: Menor ao Maior</SelectItem>
                    <SelectItem value="price-high">Preço: Maior ao Menor</SelectItem>
                    <SelectItem value="name">Nome</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          {/* Products Grid */}
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {loading ? (
              <div className="col-span-full flex flex-col items-center justify-center py-20">
                <Loader2 className="h-16 w-16 animate-spin text-[#D97757] mb-4" />
                <p className="font-body text-lg text-[#5A6751]">Carregando produtos...</p>
              </div>
            ) : error ? (
              <div className="col-span-full">
                <Card className="text-center py-16 rounded-3xl border-2 border-[#2D3319]/10">
                  <CardContent>
                    <h3 className="font-display text-2xl font-bold text-[#2D3319] mb-2">
                      Erro ao carregar produtos
                    </h3>
                    <p className="font-body text-[#5A6751] mb-6">{error}</p>
                    <Button
                      onClick={() => window.location.reload()}
                      className="btn-primary-custom font-body px-8 py-3 rounded-full"
                    >
                      Tentar Novamente
                    </Button>
                  </CardContent>
                </Card>
              </div>
            ) : filteredProducts.length === 0 ? (
              <div className="col-span-full">
                <Card className="text-center py-16 rounded-3xl border-2 border-[#2D3319]/10">
                  <CardContent>
                    <h3 className="font-display text-2xl font-bold text-[#2D3319] mb-2">
                      Nenhum produto encontrado
                    </h3>
                    <p className="font-body text-[#5A6751]">
                      Tente ajustar seus filtros ou termo de busca
                    </p>
                  </CardContent>
                </Card>
              </div>
            ) : (
              filteredProducts.map((product, index) => (
                <Card
                  key={product._id}
                  className="product-card observe-animation bg-white shadow-lg hover:shadow-2xl rounded-3xl overflow-hidden"
                  style={{ animationDelay: `${index * 0.05}s` }}
                >
                  <CardContent className="p-0">
                    <div>
                      <div className="relative aspect-square overflow-hidden group">
                        <ProductImageWithFallback
                          images={product.images || []}
                          alt={product.name}
                          className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
                          fallbackIcon="🐕"
                        />
                        <Button
                          variant="ghost"
                          size="icon"
                          className={`absolute top-3 right-3 rounded-full bg-white/90 backdrop-blur-sm shadow-lg ${
                            favorites.includes(parseInt(product._id))
                              ? "text-red-500 hover:text-red-600"
                              : "text-[#5A6751] hover:text-red-500"
                          }`}
                          onClick={() => toggleFavorite(product._id)}
                        >
                          <Heart
                            className={`h-5 w-5 ${favorites.includes(parseInt(product._id)) ? "fill-current" : ""}`}
                          />
                        </Button>
                        {product.price > 100 && (
                          <Badge className="absolute top-3 left-3 bg-[#F4C430] text-[#2D3319] font-bold rounded-full px-4 py-1">
                            15% OFF
                          </Badge>
                        )}
                        {product.quantity === 0 && (
                          <div className="absolute inset-0 bg-white/90 backdrop-blur-sm flex items-center justify-center">
                            <Badge className="bg-[#2D3319] text-white px-6 py-2 text-base">
                              Sem Estoque
                            </Badge>
                          </div>
                        )}
                      </div>

                      <div className="p-6 space-y-3">
                        {product.brand && (
                          <p className="font-body text-xs text-[#8B9A7E] font-semibold uppercase tracking-widest">
                            {product.brand}
                          </p>
                        )}
                        <Link to={`/product/${product._id}`} className="block">
                          <h3 className="font-display text-lg font-bold text-[#2D3319] hover:text-[#D97757] transition-colors line-clamp-2 min-h-[3.5rem]">
                            {product.name}
                          </h3>
                        </Link>

                        <div className="flex items-center gap-1">
                          {[1, 2, 3, 4, 5].map((star) => (
                            <Star
                              key={star}
                              className={`h-4 w-4 ${
                                star <= Math.round(product.rating || 0)
                                  ? "text-[#F4C430] fill-[#F4C430]"
                                  : "text-gray-300"
                              }`}
                            />
                          ))}
                          <span className="font-body text-sm font-semibold text-[#2D3319] ml-2">
                            {(product.rating || 0).toFixed(1)}
                          </span>
                        </div>

                        {product.category && (
                          <Badge className="bg-[#8B9A7E]/20 text-[#5A6751] border-0 font-body">
                            {product.category}
                          </Badge>
                        )}

                        <div className="space-y-3 pt-2">
                          <div className="flex items-baseline gap-2">
                            <span className="font-display text-3xl font-bold text-[#D97757]">
                              R$ {product.price.toFixed(2)}
                            </span>
                            {product.price > 100 && (
                              <span className="font-body text-sm text-[#5A6751] line-through">
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
                            onClick={() => handleAddToCart(product)}
                            disabled={product.quantity === 0}
                            className="w-full btn-primary-custom font-body rounded-full font-semibold"
                          >
                            <ShoppingCart className="h-4 w-4 mr-2" />
                            {(() => {
                              const productId = product._id || (product as any).id;
                              const quantity = getItemQuantity(productId);
                              if (quantity > 0) {
                                return `No carrinho (${quantity})`;
                              }
                              return "Adicionar";
                            })()}
                          </Button>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))
            )}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="mt-16 flex justify-center items-center gap-3">
              <Button
                variant="outline"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="font-body border-2 border-[#2D3319] hover:bg-[#2D3319] hover:text-white rounded-full px-6"
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
                          ? "bg-[#D97757] hover:bg-[#C56647] rounded-full"
                          : "border-2 border-[#2D3319]/20 rounded-full hover:border-[#D97757]"
                      }
                    >
                      {pageNum}
                    </Button>
                  );
                })}
                {totalPages > 5 && <span className="font-body text-[#5A6751]">...</span>}
              </div>
              <Button
                variant="outline"
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="font-body border-2 border-[#2D3319] hover:bg-[#2D3319] hover:text-white rounded-full px-6"
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

export default ProductList;
