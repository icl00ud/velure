import { Filter, Grid3X3, Heart, List, Loader2, Search, ShoppingCart, Star } from "lucide-react";
import { useState } from "react";
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

// No mock data needed - using real API

const ProductCatalog = () => {
  const { category } = useParams();
  const [searchQuery, setSearchQuery] = useState("");
  const [sortBy, setSortBy] = useState("popularity");
  const [filterBy, setFilterBy] = useState("all");
  const [viewMode, setViewMode] = useState("grid");
  const [favorites, setFavorites] = useState<number[]>([]);
  const [page, setPage] = useState(1);
  const pageSize = 12;

  // Use hooks para produtos e carrinho
  const { products, loading, error, totalCount, totalPages } = useProductsPaginated(
    page,
    pageSize,
    category
  );
  const { addToCart, isInCart } = useCart();

  const toggleFavorite = (productId: string) => {
    const numId = parseInt(productId);
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

  // Filtrar produtos com base na busca e filtros
  const filteredProdutos = products.filter((product) => {
    const matchesSearch =
      product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (product.brand || "").toLowerCase().includes(searchQuery.toLowerCase());
    const matchesFilter =
      filterBy === "all" ||
      (filterBy === "on-sale" && product.price) ||
      (filterBy === "in-stock" && product.disponibility);
    return matchesSearch && matchesFilter;
  });

  // Mapeamento de categorias para português
  const categoryNames: Record<string, string> = {
    dogs: "Cães",
    cats: "Gatos",
    birds: "Pássaros",
    fish: "Peixes",
    "small-pets": "Pets pequenos",
  };

  const categoryName = category ? categoryNames[category] || category : "Todos os produtos";
  const categoryEmoji =
    category === "dogs"
      ? "🐕"
      : category === "cats"
        ? "🐱"
        : category === "birds"
          ? "🦜"
          : category === "fish"
            ? "🐠"
            : category === "small-pets"
              ? "🐹"
              : "🐾";

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
            {category && (
              <>
                <span>/</span>
                <span className="text-foreground font-medium">{categoryName}</span>
              </>
            )}
          </div>
        </nav>

        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-4xl font-bold text-foreground mb-2">
                {category ? `Produtos para ${categoryName}` : "Todos os produtos"}
              </h1>
              <p className="text-muted-foreground">
                {loading ? "Carregando..." : `${totalCount} produtos encontrados`}
              </p>
            </div>
            <div className="text-6xl">{categoryEmoji}</div>
          </div>
        </div>

        {/* Filters & Search */}
        <Card className="mb-8 shadow-soft">
          <CardContent className="p-6">
            <div className="flex flex-col lg:flex-row gap-4 items-center justify-between">
              <div className="flex flex-col sm:flex-row gap-4 flex-1">
                {/* Search */}
                <div className="relative flex-1 max-w-md">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Buscar produtos..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10"
                  />
                </div>

                {/* Filter */}
                <Select value={filterBy} onValueChange={setFilterBy}>
                  <SelectTrigger className="w-40">
                    <Filter className="h-4 w-4 mr-2" />
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Todos os produtos</SelectItem>
                    <SelectItem value="on-sale">Em promoção</SelectItem>
                    <SelectItem value="in-stock">Em estoque</SelectItem>
                  </SelectContent>
                </Select>

                {/* Sort */}
                <Select value={sortBy} onValueChange={setSortBy}>
                  <SelectTrigger className="w-40">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
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

        {/* Produtos Grid */}
        <div
          className={
            viewMode === "grid"
              ? "grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6"
              : "space-y-4"
          }
        >
          {loading ? (
            // Loading state
            <div className="col-span-full flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
              <span className="ml-2 text-muted-foreground">Carregando produtos...</span>
            </div>
          ) : error ? (
            // Error state
            <div className="col-span-full">
              <Card className="text-center py-12">
                <CardContent>
                  <div className="text-6xl mb-4">⚠️</div>
                  <h3 className="text-xl font-semibold text-foreground mb-2">
                    Erro ao carregar produtos
                  </h3>
                  <p className="text-muted-foreground mb-6">{error}</p>
                  <Button
                    onClick={() => window.location.reload()}
                    className="bg-gradient-primary hover:opacity-90 text-primary-foreground"
                  >
                    Tentar novamente
                  </Button>
                </CardContent>
              </Card>
            </div>
          ) : (
            filteredProdutos.map((product) => (
              <Card
                key={product._id}
                className="group shadow-soft hover:shadow-primary transition-all duration-300"
              >
                <CardContent className="p-0">
                  {viewMode === "grid" ? (
                    // Grid View
                    <>
                      <div className="relative">
                        <div className="aspect-square">
                          <ProductImageWithFallback
                            images={product.images || []}
                            alt={product.name}
                            className="w-full h-full rounded-t-lg"
                            fallbackIcon="🐕"
                          />
                        </div>
                        {/* Placeholder for discount badge */}
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
                            <Badge variant="destructive">Sem Estoque</Badge>
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
                            <span className="text-xs font-medium ml-1">{product.rating}</span>
                          </div>
                          <span className="text-xs text-muted-foreground">(avaliações)</span>
                        </div>

                        <div className="flex flex-wrap gap-1 mb-3">
                          {product.category && (
                            <Badge variant="secondary" className="text-xs">
                              {product.category}
                            </Badge>
                          )}
                          {product.colors && product.colors.length > 0 && (
                            <Badge variant="secondary" className="text-xs">
                              {product.colors.length} cores
                            </Badge>
                          )}
                        </div>

                        <div className="flex items-center justify-between">
                          <div>
                            <div className="font-bold text-primary">
                              ${product.price.toFixed(2)}
                            </div>
                            {product.quantity_warehouse < 10 && product.quantity_warehouse > 0 && (
                              <div className="text-xs text-orange-500">
                                Apenas {product.quantity_warehouse} restantes
                              </div>
                            )}
                          </div>
                          <Button
                            size="sm"
                            className="bg-gradient-primary hover:opacity-90 text-primary-foreground"
                            onClick={() => handleAdicionarToCart(product)}
                            disabled={!product.disponibility || product.quantity_warehouse === 0}
                          >
                            <ShoppingCart className="h-3 w-3 mr-1" />
                            {isInCart(product._id) ? "Adicionado" : "Adicionar"}
                          </Button>
                        </div>
                      </div>
                    </>
                  ) : (
                    // List View
                    <div className="flex gap-4 p-4">
                      <div className="w-32 h-32 flex-shrink-0">
                        <ProductImageWithFallback
                          images={product.images || []}
                          alt={product.name}
                          className="w-full h-full rounded-lg"
                          fallbackIcon="🐕"
                        />
                      </div>

                      <div className="flex-1">
                        <div className="flex items-start justify-between mb-2">
                          <div>
                            <p className="text-sm text-muted-foreground font-medium">
                              {product.brand || "Marca"}
                            </p>
                            <Link
                              to={`/product/${product._id}`}
                              className="text-lg font-semibold text-foreground hover:text-primary transition-colors"
                            >
                              {product.name}
                            </Link>
                            {product.description && (
                              <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                                {product.description}
                              </p>
                            )}
                          </div>
                          <Button
                            variant="ghost"
                            size="icon"
                            className={
                              favorites.includes(parseInt(product._id))
                                ? "text-red-500"
                                : "text-muted-foreground"
                            }
                            onClick={() => toggleFavorite(product._id)}
                          >
                            <Heart
                              className={`h-4 w-4 ${favorites.includes(parseInt(product._id)) ? "fill-current" : ""}`}
                            />
                          </Button>
                        </div>

                        <div className="flex items-center space-x-2 mb-2">
                          <div className="flex items-center">
                            <Star className="h-4 w-4 text-accent fill-current" />
                            <span className="text-sm font-medium ml-1">{product.rating}</span>
                          </div>
                          <span className="text-sm text-muted-foreground">(avaliações)</span>
                          {!product.disponibility && (
                            <Badge variant="destructive">Sem estoque</Badge>
                          )}
                          {product.quantity_warehouse < 10 && product.quantity_warehouse > 0 && (
                            <Badge variant="outline" className="text-orange-500">
                              Estoque baixo
                            </Badge>
                          )}
                        </div>

                        <div className="flex flex-wrap gap-1 mb-3">
                          {product.category && (
                            <Badge variant="secondary" className="text-xs">
                              {product.category}
                            </Badge>
                          )}
                          {product.sku && (
                            <Badge variant="outline" className="text-xs">
                              SKU: {product.sku}
                            </Badge>
                          )}
                          {product.colors &&
                            product.colors.map((color, index) => (
                              <Badge key={index} variant="secondary" className="text-xs">
                                {color}
                              </Badge>
                            ))}
                        </div>

                        <div className="flex items-center justify-between">
                          <div className="flex items-center space-x-2">
                            <div className="text-xl font-bold text-primary">
                              ${product.price.toFixed(2)}
                            </div>
                            {product.quantity_warehouse < 10 && product.quantity_warehouse > 0 && (
                              <span className="text-sm text-orange-500">
                                Apenas {product.quantity_warehouse} restantes
                              </span>
                            )}
                          </div>
                          <Button
                            className="bg-gradient-primary hover:opacity-90 text-primary-foreground"
                            onClick={() => handleAdicionarToCart(product)}
                            disabled={!product.disponibility || product.quantity_warehouse === 0}
                          >
                            <ShoppingCart className="h-4 w-4 mr-2" />
                            {isInCart(product._id)
                              ? "Adicionado ao carrinho"
                              : "Adicionar ao carrinho"}
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
          <Card className="text-center py-12">
            <CardContent>
              <div className="text-6xl mb-4">🔍</div>
              <h3 className="text-xl font-semibold text-foreground mb-2">
                Nenhum produto encontrado
              </h3>
              <p className="text-muted-foreground mb-6">
                Tente ajustar sua busca ou critérios de filtro
              </p>
              <Button
                onClick={() => {
                  setSearchQuery("");
                  setFilterBy("all");
                }}
                className="bg-gradient-primary hover:opacity-90 text-primary-foreground"
              >
                Limpar filtros
              </Button>
            </CardContent>
          </Card>
        )}
      </main>
    </div>
  );
};

export default ProductCatalog;
