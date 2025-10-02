import { Filter, Grid3X3, Heart, List, Loader2, Search, ShoppingCart, Star } from "lucide-react";
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
  const [viewMode, setViewMode] = useState("grid");
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
  const { addToCart, isInCart } = useCart();

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

  // Filtrar produtos com base na busca
  const filteredProducts = products.filter(
    (product) =>
      product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (product.brand || "").toLowerCase().includes(searchQuery.toLowerCase())
  );

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
            <span className="text-foreground font-medium">Todos os Produtos</span>
          </div>
        </nav>

        {/* Header Section */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-4xl font-bold text-foreground mb-2">Todos os Produtos</h1>
              <p className="text-muted-foreground">
                {loading ? "Carregando..." : `${totalCount} produtos disponíveis`}
              </p>
            </div>
            <div className="text-6xl">🐾</div>
          </div>
        </div>

        {/* Filters */}
        <Card className="mb-6 shadow-soft">
          <CardContent className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
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

              {/* View Mode */}
              <div className="flex gap-2">
                <Button
                  variant={viewMode === "grid" ? "default" : "outline"}
                  size="icon"
                  onClick={() => setViewMode("grid")}
                >
                  <Grid3X3 className="h-4 w-4" />
                </Button>
                <Button
                  variant={viewMode === "list" ? "default" : "outline"}
                  size="icon"
                  onClick={() => setViewMode("list")}
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
              : "space-y-4"
          }
        >
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
                  <div className="text-6xl mb-4">⚠️</div>
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
                  <div className="text-6xl mb-4">🔍</div>
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
                className="group shadow-soft hover:shadow-primary transition-all duration-300"
              >
                <CardContent className="p-0">
                  <div>
                    <div className="relative aspect-square overflow-hidden rounded-t-lg bg-muted">
                      <ProductImageWithFallback
                        images={product.images || []}
                        alt={product.name}
                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                        fallbackIcon="🐕"
                      />
                      <Button
                        variant="ghost"
                        size="icon"
                        className={`absolute top-2 right-2 ${
                          favorites.includes(parseInt(product._id))
                            ? "text-red-500 hover:text-red-600"
                            : "text-muted-foreground hover:text-red-500"
                        }`}
                        onClick={() => toggleFavorite(product._id)}
                      >
                        <Heart
                          className={`h-4 w-4 ${favorites.includes(parseInt(product._id)) ? "fill-current" : ""}`}
                        />
                      </Button>
                      {!product.disponibility && (
                        <div className="absolute inset-0 bg-background/80 rounded-t-lg flex items-center justify-center">
                          <Badge variant="secondary">Sem Estoque</Badge>
                        </div>
                      )}
                    </div>

                    <div className="p-4">
                      <div className="mb-2">
                        <p className="text-xs text-muted-foreground font-medium">
                          {product.brand || "Marca"}
                        </p>
                        <Link
                          to={`/product/${product._id}`}
                          className="font-semibold text-foreground hover:text-primary transition-colors line-clamp-2"
                        >
                          {product.name}
                        </Link>
                      </div>

                      <div className="flex items-center space-x-1 mb-2">
                        <div className="flex items-center">
                          <Star className="h-3 w-3 text-accent fill-current" />
                          <Star className="h-3 w-3 text-accent fill-current" />
                          <Star className="h-3 w-3 text-accent fill-current" />
                          <Star className="h-3 w-3 text-accent fill-current" />
                          <Star className="h-3 w-3 text-muted-foreground" />
                        </div>
                        <span className="text-xs text-muted-foreground">(4.0)</span>
                      </div>

                      <div className="flex flex-wrap gap-1 mb-2">
                        {product.category && (
                          <Badge variant="secondary" className="text-xs">
                            {product.category}
                          </Badge>
                        )}
                      </div>

                      <div className="flex items-center justify-between">
                        <div>
                          <div className="font-bold text-primary">${product.price.toFixed(2)}</div>
                          {product.quantity_warehouse < 10 && product.quantity_warehouse > 0 && (
                            <div className="text-xs text-orange-500">
                              Apenas {product.quantity_warehouse} restantes
                            </div>
                          )}
                        </div>
                        <Button
                          size="sm"
                          onClick={() => handleAddToCart(product)}
                          disabled={!product.disponibility || isInCart(product._id)}
                          className="bg-gradient-primary hover:opacity-90 text-primary-foreground"
                        >
                          <ShoppingCart className="h-3 w-3 mr-1" />
                          {isInCart(product._id) ? "No Carrinho" : "Adicionar"}
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
          <div className="mt-8 flex justify-center gap-2">
            <Button
              variant="outline"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              Anterior
            </Button>
            <div className="flex items-center gap-2">
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                const pageNum = i + 1;
                return (
                  <Button
                    key={pageNum}
                    variant={page === pageNum ? "default" : "outline"}
                    onClick={() => setPage(pageNum)}
                  >
                    {pageNum}
                  </Button>
                );
              })}
            </div>
            <Button
              variant="outline"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
            >
              Próximo
            </Button>
          </div>
        )}
      </main>
    </div>
  );
};

export default ProductList;
