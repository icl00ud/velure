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
      <div className="min-h-screen bg-gradient-to-br from-[#F8FAF5] via-[#EDF7ED] to-[#E8F5E9] flex items-center justify-center p-4 grain-texture relative overflow-hidden">
        {/* Decorative Elements */}
        <div className="fixed top-20 right-10 w-64 h-64 rounded-full bg-[#52B788]/10 blur-3xl pointer-events-none" />
        <div className="fixed bottom-20 left-10 w-80 h-80 rounded-full bg-[#95D5B2]/10 blur-3xl pointer-events-none" />

        <div
          className={`text-center relative z-10 max-w-2xl mx-auto ${isVisible ? "hero-enter active" : "hero-enter"}`}
        >
          {/* 404 Illustration */}
          <div className="relative inline-block mb-12">
            <div className="absolute inset-0 bg-[#52B788]/20 blur-3xl" />
            <div className="relative">
              <h1 className="font-display text-[180px] lg:text-[240px] font-bold text-[#52B788] leading-none opacity-20">
                404
              </h1>
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="bg-gradient-to-br from-[#52B788] to-[#40916C] rounded-full p-8 shadow-2xl">
                  <Search className="h-16 w-16 text-white" />
                </div>
              </div>
            </div>
          </div>

          {/* Content */}
          <div className="space-y-6">
            <h2 className="font-display text-4xl lg:text-5xl font-bold text-[#1B4332]">
              Ops! Página não encontrada
            </h2>
            <div className="w-20 h-1 bg-gradient-to-r from-[#52B788] to-[#A7C957] mx-auto" />
            <p className="font-body text-xl text-[#2D6A4F] max-w-lg mx-auto">
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
              className="font-body text-lg rounded-full px-8 py-6 border-2 border-[#1B4332] hover:bg-[#1B4332] hover:text-white"
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
