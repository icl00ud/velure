import { Home, Search } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { designSystemStyles } from "@/styles/design-system";

const NotFound = () => {
  const location = useLocation();
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    setIsVisible(true);
    console.error("404 Error: User attempted to access non-existent route:", location.pathname);
  }, [location.pathname]);

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-gradient-to-br from-[#FAF7F2] via-[#F5EFE7] to-[#EDDBC7] flex items-center justify-center p-4 grain-texture relative overflow-hidden">
        {/* Decorative Elements */}
        <div className="fixed top-20 right-10 w-64 h-64 rounded-full bg-[#D97757]/10 blur-3xl pointer-events-none" />
        <div className="fixed bottom-20 left-10 w-80 h-80 rounded-full bg-[#8B9A7E]/10 blur-3xl pointer-events-none" />

        <div className={`text-center relative z-10 max-w-2xl mx-auto ${isVisible ? 'hero-enter active' : 'hero-enter'}`}>
          {/* 404 Illustration */}
          <div className="relative inline-block mb-12">
            <div className="absolute inset-0 bg-[#D97757]/20 blur-3xl" />
            <div className="relative">
              <h1 className="font-display text-[180px] lg:text-[240px] font-bold text-[#D97757] leading-none opacity-20">
                404
              </h1>
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="bg-gradient-to-br from-[#D97757] to-[#C56647] rounded-full p-8 shadow-2xl">
                  <Search className="h-16 w-16 text-white" />
                </div>
              </div>
            </div>
          </div>

          {/* Content */}
          <div className="space-y-6">
            <h2 className="font-display text-4xl lg:text-5xl font-bold text-[#2D3319]">
              Ops! Página não encontrada
            </h2>
            <div className="w-20 h-1 bg-gradient-to-r from-[#D97757] to-[#F4C430] mx-auto" />
            <p className="font-body text-xl text-[#5A6751] max-w-lg mx-auto">
              Parece que você se perdeu! A página que você está procurando não existe ou foi movida.
            </p>
          </div>

          {/* Actions */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center mt-12">
            <Button asChild className="btn-primary-custom font-body text-lg rounded-full px-8 py-6">
              <Link to="/">
                <Home className="h-5 w-5 mr-2" />
                Voltar ao início
              </Link>
            </Button>
            <Button
              asChild
              variant="outline"
              className="font-body text-lg rounded-full px-8 py-6 border-2 border-[#2D3319] hover:bg-[#2D3319] hover:text-white"
            >
              <Link to="/products">
                <Search className="h-5 w-5 mr-2" />
                Explorar produtos
              </Link>
            </Button>
          </div>

          {/* Decorative pawprints */}
          <div className="mt-16 flex justify-center space-x-4 opacity-20">
            <span className="text-4xl">🐾</span>
            <span className="text-3xl">🐾</span>
            <span className="text-4xl">🐾</span>
          </div>
        </div>
      </div>
    </>
  );
};

export default NotFound;
