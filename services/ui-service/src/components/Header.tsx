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
import { designSystemStyles } from "@/styles/design-system";

const Header = () => {
  const { itemsCount } = useCart();
  const { isAuthenticated, logout } = useAuth();
  const [categories, setCategories] = useState<string[]>([]);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);

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

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 20);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
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
    <>
      <style>{designSystemStyles}</style>
      <header
        className={`sticky top-0 z-50 w-full transition-all duration-300 ${
          scrolled
            ? "bg-white/95 backdrop-blur-md shadow-lg border-b-2 border-[#D97757]/20"
            : "bg-[#FAF7F2]/95 backdrop-blur-sm border-b border-[#2D3319]/10"
        }`}
      >
        <div className="container mx-auto px-4 lg:px-8">
          <div className="flex items-center justify-between h-20">
            {/* Logo */}
            <Link to="/" className="flex items-center space-x-3 group">
              <div className="relative">
                <div className="absolute inset-0 bg-gradient-to-br from-[#D97757] to-[#C56647] rounded-2xl blur-sm group-hover:blur-md transition-all opacity-50" />
                <div className="relative bg-gradient-to-br from-[#D97757] to-[#C56647] rounded-2xl p-2.5 transform group-hover:scale-110 transition-transform duration-300">
                  <Heart className="h-6 w-6 text-white fill-white" />
                </div>
              </div>
              <span className="font-display font-bold text-2xl text-[#2D3319] group-hover:text-[#D97757] transition-colors">
                Velure
              </span>
            </Link>

            {/* Navigation - Desktop */}
            <nav className="hidden md:flex items-center space-x-8">
              <div className="flex items-center">
                <Link
                  to="/products"
                  className="font-body text-[#2D3319] hover:text-[#D97757] transition-colors font-medium relative group"
                >
                  Produtos
                  <span className="absolute -bottom-1 left-0 w-0 h-0.5 bg-[#D97757] group-hover:w-full transition-all duration-300" />
                </Link>
                {categories.length > 0 && (
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 ml-1 hover:bg-[#D97757]/10 hover:text-[#D97757]"
                      >
                        <ChevronDown className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent className="w-56 bg-white border-2 border-[#2D3319]/10 shadow-2xl rounded-2xl p-2">
                      {categories.map((category) => (
                        <DropdownMenuItem
                          key={category}
                          asChild
                          className="rounded-xl font-body hover:bg-[#D97757]/10 hover:text-[#D97757] cursor-pointer"
                        >
                          <Link to={`/products/${category}`} className="w-full px-3 py-2">
                            {formatCategoryName(category)}
                          </Link>
                        </DropdownMenuItem>
                      ))}
                    </DropdownMenuContent>
                  </DropdownMenu>
                )}
              </div>

              <Link
                to="/contact"
                className="font-body text-[#2D3319] hover:text-[#D97757] transition-colors font-medium relative group"
              >
                Contato
                <span className="absolute -bottom-1 left-0 w-0 h-0.5 bg-[#D97757] group-hover:w-full transition-all duration-300" />
              </Link>
            </nav>

            {/* Actions */}
            <div className="flex items-center space-x-2">
              {/* Mobile Menu */}
              <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
                <SheetTrigger asChild className="md:hidden">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="hover:bg-[#D97757]/10 hover:text-[#D97757]"
                  >
                    <Menu className="h-5 w-5" />
                  </Button>
                </SheetTrigger>
                <SheetContent side="left" className="w-[280px] sm:w-[320px] bg-[#FAF7F2]">
                  <SheetHeader>
                    <SheetTitle className="flex items-center space-x-3">
                      <div className="bg-gradient-to-br from-[#D97757] to-[#C56647] rounded-2xl p-2">
                        <Heart className="h-5 w-5 text-white fill-white" />
                      </div>
                      <span className="font-display font-bold text-xl text-[#2D3319]">Velure</span>
                    </SheetTitle>
                  </SheetHeader>
                  <nav className="flex flex-col space-y-2 mt-8 font-body">
                    <Link
                      to="/products"
                      className="text-[#2D3319] hover:text-[#D97757] hover:bg-[#D97757]/10 transition-colors py-3 px-4 rounded-xl border-b border-[#2D3319]/10"
                      onClick={() => setMobileMenuOpen(false)}
                    >
                      Todos os Produtos
                    </Link>
                    {categories.length > 0 && (
                      <div className="space-y-1 pt-2">
                        <p className="text-sm font-semibold text-[#5A6751] px-4 mb-2">
                          Categorias
                        </p>
                        {categories.map((category) => (
                          <Link
                            key={category}
                            to={`/products/${category}`}
                            className="block text-[#2D3319] hover:text-[#D97757] hover:bg-[#D97757]/10 transition-colors py-2 px-6 rounded-xl"
                            onClick={() => setMobileMenuOpen(false)}
                          >
                            {formatCategoryName(category)}
                          </Link>
                        ))}
                      </div>
                    )}
                    <Link
                      to="/contact"
                      className="text-[#2D3319] hover:text-[#D97757] hover:bg-[#D97757]/10 transition-colors py-3 px-4 rounded-xl border-t border-[#2D3319]/10 mt-2"
                      onClick={() => setMobileMenuOpen(false)}
                    >
                      Contato
                    </Link>
                    {isAuthenticated ? (
                      <>
                        <Link
                          to="/orders"
                          className="text-[#2D3319] hover:text-[#D97757] hover:bg-[#D97757]/10 transition-colors py-3 px-4 rounded-xl"
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
                          className="text-left text-[#2D3319] hover:text-[#D97757] hover:bg-[#D97757]/10 transition-colors py-3 px-4 rounded-xl"
                        >
                          Sair
                        </button>
                      </>
                    ) : (
                      <Link
                        to="/login"
                        className="text-[#2D3319] hover:text-[#D97757] hover:bg-[#D97757]/10 transition-colors py-3 px-4 rounded-xl"
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
                <Button
                  variant="ghost"
                  size="icon"
                  className="relative hover:bg-[#D97757]/10 hover:text-[#D97757] transition-all"
                >
                  <ShoppingCart className="h-5 w-5" />
                  {itemsCount > 0 && (
                    <span className="absolute -top-1 -right-1 bg-gradient-to-br from-[#D97757] to-[#C56647] text-white text-xs font-bold rounded-full h-5 w-5 flex items-center justify-center shadow-lg animate-pulse">
                      {itemsCount}
                    </span>
                  )}
                </Button>
              </Link>

              {/* User */}
              {isAuthenticated ? (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="hover:bg-[#D97757]/10 hover:text-[#D97757]"
                    >
                      <User className="h-5 w-5" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent className="w-56 bg-white border-2 border-[#2D3319]/10 shadow-2xl rounded-2xl p-2">
                    <DropdownMenuItem
                      asChild
                      className="rounded-xl font-body hover:bg-[#D97757]/10 hover:text-[#D97757] cursor-pointer"
                    >
                      <Link to="/orders" className="w-full px-3 py-2">
                        <ShoppingCart className="h-4 w-4 mr-2" />
                        Meus Pedidos
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={handleLogout}
                      className="rounded-xl font-body hover:bg-[#D97757]/10 hover:text-[#D97757] cursor-pointer px-3 py-2"
                    >
                      <LogOut className="h-4 w-4 mr-2" />
                      Sair
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              ) : (
                <Link to="/login">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="hover:bg-[#D97757]/10 hover:text-[#D97757]"
                  >
                    <User className="h-5 w-5" />
                  </Button>
                </Link>
              )}
            </div>
          </div>
        </div>
      </header>
    </>
  );
};

export default Header;
