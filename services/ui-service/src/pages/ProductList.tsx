import { Heart, Loader2, Search, ShoppingCart, Star } from "lucide-react";
import { useState } from "react";
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

const ProductList = () => {
  const [searchQuery, setSearchQuery] = useState("");
  const [sortBy, setSortBy] = useState("popularity");
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [favorites, setFavorites] = useState<number[]>([]);
  const [page, setPage] = useState(1);
  const pageSize = 12;

  // Hooks
  const { categories, loading: loadingCategories } = useCategories();
  const { products, loading, error, totalCount, totalPages } = useProductsPaginated(
    page,
    pageSize,
    selectedCategory !== "all" ? selectedCategory : undefined
  );
  const { addToCart, isInCart, getItemQuantity } = useCart();

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

  // Filtrar e ordenar produtos
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
          return 0; // Mantém ordem original (do banco)
      }
    });

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
            <span className="text-foreground font-medium">Todos os produtos</span>
          </div>
        </nav>

        {/* Header Section */}
        <div className="mb-8 text-center lg:text-left">
          <h1 className="text-4xl lg:text-5xl font-bold text-foreground mb-3">
            Todos os produtos
          </h1>
          <p className="text-lg text-muted-foreground">
            {loading ? (
              <span className="inline-flex items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                Carregando produtos...
              </span>
            ) : (
              <>
                <span className="font-semibold text-primary">{totalCount || 0}</span> produtos disponíveis para seu pet
              </>
            )}
          </p>
        </div>

        {/* Filters */}
        <Card className="mb-6 shadow-soft">
          <CardContent className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {/* Search */}
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Buscar produtos..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>

              {/* Category Filter */}
              <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                <SelectTrigger>
                  <SelectValue placeholder="Todas as Categorias" />
                </SelectTrigger>
                <SelectContent>
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
                <SelectTrigger>
                  <SelectValue placeholder="Ordenar por" />
                </SelectTrigger>
                <SelectContent>
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
            <div className="col-span-full flex items-center justify-center py-12">
              <div className="text-center">
                <Loader2 className="h-12 w-12 animate-spin text-primary mx-auto mb-4" />
                <p className="text-muted-foreground">Carregando produtos...</p>
              </div>
            </div>
          ) : error ? (
            <div className="col-span-full">
              <Card className="text-center py-12">
                <CardContent>
                  <h3 className="text-xl font-semibold text-foreground mb-2">
                    Erro ao carregar produtos
                  </h3>
                  <p className="text-muted-foreground mb-6">{error}</p>
                  <Button onClick={() => window.location.reload()} variant="outline">
                    Tentar Novamente
                  </Button>
                </CardContent>
              </Card>
            </div>
          ) : filteredProducts.length === 0 ? (
            <div className="col-span-full">
              <Card className="text-center py-12">
                <CardContent>
                  <h3 className="text-xl font-semibold text-foreground mb-2">
                    Nenhum produto encontrado
                  </h3>
                  <p className="text-muted-foreground">
                    Tente ajustar seus filtros ou termo de busca
                  </p>
                </CardContent>
              </Card>
            </div>
          ) : (
            filteredProducts.map((product) => (
              <Card
                key={product._id}
                className="group shadow-soft hover:shadow-lg hover:-translate-y-1 transition-all duration-300 border border-transparent hover:border-primary/20"
              >
                <CardContent className="p-0">
                  <div>
                    <div className="relative aspect-square overflow-hidden rounded-t-lg bg-gradient-to-br from-muted to-muted/50">
                      <ProductImageWithFallback
                        images={product.images || []}
                        alt={product.name}
                        className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-500"
                        fallbackIcon="🐕"
                      />
                      <Button
                        variant="ghost"
                        size="icon"
                        className={`absolute top-3 right-3 backdrop-blur-sm bg-background/80 rounded-full ${
                          favorites.includes(parseInt(product._id))
                            ? "text-red-500 hover:text-red-600 hover:bg-red-50"
                            : "text-muted-foreground hover:text-red-500 hover:bg-background"
                        } transition-all duration-200`}
                        onClick={() => toggleFavorite(product._id)}
                      >
                        <Heart
                          className={`h-5 w-5 ${favorites.includes(parseInt(product._id)) ? "fill-current" : ""}`}
                        />
                      </Button>
                      {product.price > 100 && (
                        <Badge className="absolute top-3 left-3 bg-secondary text-secondary-foreground font-semibold">
                          15% OFF
                        </Badge>
                      )}
                      {!product.disponibility && (
                        <div className="absolute inset-0 bg-background/90 backdrop-blur-sm rounded-t-lg flex items-center justify-center">
                          <Badge variant="secondary" className="text-base px-4 py-2">Sem Estoque</Badge>
                        </div>
                      )}
                    </div>

                    <div className="p-4 space-y-2">
                      {product.brand && (
                        <p className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                          {product.brand}
                        </p>
                      )}
                      <Link
                        to={`/product/${product._id}`}
                        className="block"
                      >
                        <h3 className="font-semibold text-foreground hover:text-primary transition-colors line-clamp-2 min-h-[2.5rem]">
                          {product.name}
                        </h3>
                      </Link>

                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-1">
                          {[1, 2, 3, 4, 5].map((star) => (
                            <Star 
                              key={star} 
                              className={`h-4 w-4 ${
                                star <= Math.round(product.rating || 0)
                                  ? "text-yellow-400 fill-yellow-400"
                                  : "text-gray-300"
                              }`}
                            />
                          ))}
                          <span className="text-sm font-medium text-foreground ml-1">
                            {(product.rating || 0).toFixed(1)}
                          </span>
                        </div>
                      </div>

                      {product.category && (
                        <Badge variant="outline" className="text-xs font-medium border-primary/20 text-primary">
                          {product.category}
                        </Badge>
                      )}

                      <div className="space-y-3">
                        <div className="flex items-baseline gap-2">
                          <span className="text-2xl font-bold text-primary">R$ {product.price.toFixed(2)}</span>
                          {product.price > 100 && (
                            <span className="text-xs text-muted-foreground line-through">
                              R$ {(product.price * 1.15).toFixed(2)}
                            </span>
                          )}
                        </div>
                        {product.quantity_warehouse < 10 && product.quantity_warehouse > 0 && (
                          <div className="flex items-center gap-1">
                            <div className="w-2 h-2 bg-orange-500 rounded-full animate-pulse"></div>
                            <span className="text-xs font-medium text-orange-600">
                              Apenas {product.quantity_warehouse} em estoque
                            </span>
                          </div>
                        )}
                        <Button
                          size="sm"
                          onClick={() => handleAddToCart(product)}
                          disabled={!product.disponibility}
                          className="w-full bg-gradient-primary hover:opacity-90 text-primary-foreground font-medium"
                        >
                          <ShoppingCart className="h-4 w-4 mr-2" />
                          {(() => {
                            const productId = product._id || (product as any).id;
                            const quantity = getItemQuantity(productId);
                            if (quantity > 0) {
                              return `No carrinho (${quantity})`;
                            }
                            return "Adicionar ao carrinho";
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
          <div className="mt-12 flex justify-center items-center gap-2">
            <Button
              variant="outline"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="hover:bg-primary hover:text-primary-foreground"
            >
              ← Anterior
            </Button>
            <div className="flex items-center gap-1">
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                const pageNum = i + 1;
                return (
                  <Button
                    key={pageNum}
                    variant={page === pageNum ? "default" : "ghost"}
                    onClick={() => setPage(pageNum)}
                    className={page === pageNum ? "bg-primary text-primary-foreground" : ""}
                  >
                    {pageNum}
                  </Button>
                );
              })}
              {totalPages > 5 && <span className="px-2 text-muted-foreground">...</span>}
            </div>
            <Button
              variant="outline"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="hover:bg-primary hover:text-primary-foreground"
            >
              Próximo →
            </Button>
          </div>
        )}
      </main>
    </div>
  );
};

export default ProductList;
