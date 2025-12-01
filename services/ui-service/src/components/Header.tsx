import { ChevronDown, Heart, LogOut, Menu, ShoppingCart, User } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { useAuth } from "@/hooks/use-auth";
import { useCart } from "@/hooks/use-cart";
import { productService } from "@/services/product.service";

const Header = () => {
  const { itemsCount } = useCart();
  const { isAuthenticated, logout } = useAuth();
  const [categories, setCategories] = useState<string[]>([]);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  useEffect(() => {
    const loadCategories = async () => {
      try {
        const data = await productService.getCategories();
        setCategories(data);
      } catch (error) {
        console.error("Failed to load categories:", error);
      }
    };
    loadCategories();
  }, []);

  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error("Logout failed:", error);
    }
  };

  const formatCategoryName = (category: string): string => {
    const nameMap: Record<string, string> = {
      dogs: "Cães",
      cats: "Gatos",
      birds: "Pássaros",
      fish: "Peixes",
      "small-pets": "Pets pequenos",
      reptiles: "Répteis",
      rabbits: "Coelhos",
    };
    return nameMap[category.toLowerCase()] || category;
  };

  return (
    <header className="sticky top-0 z-50 w-full bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-b border-border">
      <div className="container mx-auto px-4 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link to="/" className="flex items-center space-x-2">
          <div className="bg-gradient-primary rounded-full p-2">
            <Heart className="h-6 w-6 text-primary-foreground" />
          </div>
          <span className="font-bold text-xl text-primary">Velure</span>
        </Link>

        {/* Navigation */}
        <nav className="hidden md:flex items-center space-x-8">
          <div className="flex items-center">
            <Link to="/products" className="text-foreground hover:text-primary transition-colors">
              Produtos
            </Link>
            {categories.length > 0 && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon" className="h-8 w-8 ml-1">
                    <ChevronDown className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-56 bg-background border border-border shadow-soft">
                  {categories.map((category) => (
                    <DropdownMenuItem key={category} asChild>
                      <Link to={`/products/${category}`} className="w-full">
                        {formatCategoryName(category)}
                      </Link>
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>

          <Link to="/contact" className="text-foreground hover:text-primary transition-colors">
            Contato
          </Link>
        </nav>

        {/* Actions */}
        <div className="flex items-center space-x-4">
          {/* Mobile Menu */}
          <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
            <SheetTrigger asChild className="md:hidden">
              <Button variant="ghost" size="icon">
                <Menu className="h-5 w-5" />
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="w-[280px] sm:w-[320px]">
              <SheetHeader>
                <SheetTitle className="flex items-center space-x-2">
                  <div className="bg-gradient-primary rounded-full p-2">
                    <Heart className="h-5 w-5 text-primary-foreground" />
                  </div>
                  <span className="font-bold text-lg text-primary">Velure</span>
                </SheetTitle>
              </SheetHeader>
              <nav className="flex flex-col space-y-4 mt-8">
                <Link
                  to="/products"
                  className="text-foreground hover:text-primary transition-colors py-2 border-b border-border"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  Todos os Produtos
                </Link>
                {categories.length > 0 && (
                  <div className="space-y-2">
                    <p className="text-sm font-medium text-muted-foreground">Categorias</p>
                    {categories.map((category) => (
                      <Link
                        key={category}
                        to={`/products/${category}`}
                        className="block text-foreground hover:text-primary transition-colors py-2 pl-4"
                        onClick={() => setMobileMenuOpen(false)}
                      >
                        {formatCategoryName(category)}
                      </Link>
                    ))}
                  </div>
                )}
                <Link
                  to="/contact"
                  className="text-foreground hover:text-primary transition-colors py-2 border-t border-border"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  Contato
                </Link>
                {isAuthenticated ? (
                  <>
                    <Link
                      to="/orders"
                      className="text-foreground hover:text-primary transition-colors py-2"
                      onClick={() => setMobileMenuOpen(false)}
                    >
                      Meus Pedidos
                    </Link>
                    <button
                      type="button"
                      onClick={() => {
                        handleLogout();
                        setMobileMenuOpen(false);
                      }}
                      className="text-left text-foreground hover:text-primary transition-colors py-2"
                    >
                      Sair
                    </button>
                  </>
                ) : (
                  <Link
                    to="/login"
                    className="text-foreground hover:text-primary transition-colors py-2"
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    Entrar / Cadastrar
                  </Link>
                )}
              </nav>
            </SheetContent>
          </Sheet>

          {/* Cart */}
          <Link to="/cart">
            <Button variant="ghost" size="icon" className="relative">
              <ShoppingCart className="h-5 w-5" />
              {itemsCount > 0 && (
                <span className="absolute -top-2 -right-2 bg-secondary text-secondary-foreground text-xs rounded-full h-5 w-5 flex items-center justify-center">
                  {itemsCount}
                </span>
              )}
            </Button>
          </Link>

          {/* User */}
          {isAuthenticated ? (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon">
                  <User className="h-5 w-5" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56 bg-background border border-border shadow-soft">
                <DropdownMenuItem asChild>
                  <Link to="/orders" className="w-full">
                    <ShoppingCart className="h-4 w-4 mr-2" />
                    Meus Pedidos
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem onClick={handleLogout}>
                  <LogOut className="h-4 w-4 mr-2" />
                  Sair
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <Link to="/login">
              <Button variant="ghost" size="icon">
                <User className="h-5 w-5" />
              </Button>
            </Link>
          )}
        </div>
      </div>
    </header>
  );
};

export default Header;
