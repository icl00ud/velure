import { ArrowLeft, Loader2, Minus, Plus, Shield, ShoppingBag, Trash2, Truck } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import Header from "@/components/Header";
import { ProductImageWithFallback } from "@/components/ProductImage";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { useCart } from "@/hooks/use-cart";
import { toast } from "@/hooks/use-toast";
import { orderService } from "@/services/order.service";
import { designSystemStyles } from "@/styles/design-system";

const Cart = () => {
  const navigate = useNavigate();
  const { cartItems, updateQuantity, removeFromCart, clearCart, totalPrice } = useCart();
  const [promoCode, setPromoCode] = useState("");
  const [appliedDiscount, setAppliedDiscount] = useState(0);
  const [isProcessing, setIsProcessing] = useState(false);
  const [isVisible, setIsVisible] = useState(false);

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
  }, [cartItems]);

  const handleRemoveItem = (productId: string, productName: string) => {
    removeFromCart(productId);
    toast({
      title: "Item removido",
      description: `${productName} foi removido do seu carrinho.`,
    });
  };

  const applyPromoCode = () => {
    const code = promoCode.toUpperCase().trim();
    if (code === "VELURE10") {
      const discount = totalPrice * 0.1;
      setAppliedDiscount(discount);
      toast({
        title: "Cupom aplicado!",
        description: `Você economizou R$ ${discount.toFixed(2)} com o código ${code}`,
      });
      setPromoCode("");
    } else if (code === "VELURE5") {
      setAppliedDiscount(5.0);
      toast({
        title: "Cupom aplicado!",
        description: "Você economizou R$ 5,00 com o código VELURE5",
      });
      setPromoCode("");
    } else {
      toast({
        title: "Cupom inválido",
        description: "O código promocional não é válido.",
        variant: "destructive",
      });
    }
  };

  const handleCheckout = async () => {
    const tokenString = localStorage.getItem("token");
    if (!tokenString) {
      toast({
        title: "Autenticação necessária",
        description: "Você precisa estar logado para finalizar a compra",
        variant: "destructive",
      });
      navigate("/login");
      return;
    }

    setIsProcessing(true);
    try {
      const order = await orderService.createOrder(cartItems);

      toast({
        title: "Pedido criado com sucesso!",
        description: `Seu pedido #${order.order_id} foi criado. Total: R$ ${order.total.toFixed(2)}`,
      });

      clearCart();
      navigate("/orders");
    } catch (error) {
      toast({
        title: "Erro ao processar pedido",
        description: error instanceof Error ? error.message : "Tente novamente mais tarde",
        variant: "destructive",
      });
    } finally {
      setIsProcessing(false);
    }
  };

  const subtotal = totalPrice;
  const shipping = subtotal > 100 ? 0 : 15.0;
  const tax = subtotal * 0.05;
  const total = subtotal + shipping + tax - appliedDiscount;

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-[#F8FAF5]">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-12">
          <div className={`mb-12 ${isVisible ? 'hero-enter active' : 'hero-enter'}`}>
            <Link
              to="/"
              className="inline-flex items-center font-body text-[#2D6A4F] hover:text-[#52B788] transition-colors mb-6 group"
            >
              <ArrowLeft className="h-5 w-5 mr-2 group-hover:-translate-x-1 transition-transform" />
              Continuar comprando
            </Link>
            <h1 className="font-display text-5xl lg:text-6xl font-bold text-[#1B4332] mb-4">
              Carrinho de compras
            </h1>
            <div className="w-20 h-1 bg-gradient-to-r from-[#52B788] to-[#A7C957] mb-6" />
            <p className="font-body text-xl text-[#2D6A4F]">
              <span className="font-bold text-[#52B788]">{cartItems.length}</span>{" "}
              {cartItems.length === 1 ? "item" : "itens"} no seu carrinho
            </p>
          </div>

          {cartItems.length === 0 ? (
            <Card className="text-center py-20 rounded-3xl border-2 border-[#1B4332]/10 shadow-2xl">
              <CardContent>
                <div className="relative inline-block mb-8">
                  <div className="absolute inset-0 bg-[#52B788]/20 blur-3xl" />
                  <div className="relative bg-gradient-to-br from-[#52B788]/10 to-[#95D5B2]/10 rounded-full p-8">
                    <ShoppingBag className="h-20 w-20 text-[#52B788]" />
                  </div>
                </div>
                <h3 className="font-display text-3xl font-bold text-[#1B4332] mb-4">
                  Seu carrinho está vazio
                </h3>
                <p className="font-body text-lg text-[#2D6A4F] mb-8 max-w-md mx-auto">
                  Parece que você ainda não adicionou nenhum item ao seu carrinho.
                  Explore nossos produtos!
                </p>
                <Button
                  asChild
                  className="btn-primary-custom font-body px-10 py-6 rounded-full text-lg"
                >
                  <Link to="/products">Começar a comprar</Link>
                </Button>
              </CardContent>
            </Card>
          ) : (
            <div className="grid lg:grid-cols-3 gap-8">
              {/* Cart Items */}
              <div className="lg:col-span-2 space-y-6">
                {cartItems.map((item, index) => (
                  <Card
                    key={item.product._id}
                    className="shadow-lg border-2 border-[#1B4332]/10 rounded-3xl card-hover-subtle observe-animation"
                    style={{ animationDelay: `${index * 0.1}s` }}
                  >
                    <CardContent className="p-6">
                      <div className="flex gap-6">
                        <div className="w-32 h-32 rounded-2xl overflow-hidden flex-shrink-0 bg-gradient-to-br from-[#F8FAF5] to-white">
                          <ProductImageWithFallback
                            images={item.product.images || []}
                            alt={item.product.name}
                            className="w-full h-full object-cover hover:scale-110 transition-transform duration-300"
                            fallbackIcon="🐕"
                          />
                        </div>

                        <div className="flex-1">
                          <div className="flex justify-between items-start mb-3">
                            <div>
                              <Link
                                to={`/product/${item.product._id}`}
                                className="font-display text-xl font-bold text-[#1B4332] hover:text-[#52B788] transition-colors"
                              >
                                {item.product.name}
                              </Link>
                              <p className="font-body text-sm text-[#2D6A4F] mt-1">
                                {item.product.brand || "Marca"} •{" "}
                                {item.product.category || "Categoria"}
                              </p>
                              {item.product.quantity === 0 && (
                                <p className="font-body text-sm text-red-600 font-semibold mt-2">
                                  ⚠️ Sem estoque
                                </p>
                              )}
                            </div>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleRemoveItem(item.product._id, item.product.name)}
                              className="text-[#2D6A4F] hover:text-red-600 hover:bg-red-50 rounded-full"
                            >
                              <Trash2 className="h-5 w-5" />
                            </Button>
                          </div>

                          <div className="flex items-center justify-between">
                            <div className="flex items-center space-x-3 bg-[#F8FAF5] rounded-full p-1">
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 rounded-full hover:bg-[#52B788] hover:text-white"
                                onClick={() => updateQuantity(item.product._id, item.quantity - 1)}
                                disabled={item.quantity <= 1}
                              >
                                <Minus className="h-4 w-4" />
                              </Button>
                              <span className="w-12 text-center font-body font-bold text-[#1B4332]">
                                {item.quantity}
                              </span>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 rounded-full hover:bg-[#52B788] hover:text-white"
                                onClick={() => updateQuantity(item.product._id, item.quantity + 1)}
                              >
                                <Plus className="h-4 w-4" />
                              </Button>
                            </div>

                            <div className="text-right">
                              <div className="font-display text-2xl font-bold text-[#52B788]">
                                R$ {(item.product.price * item.quantity).toFixed(2)}
                              </div>
                              <div className="font-body text-sm text-[#2D6A4F]">
                                R$ {item.product.price.toFixed(2)} cada
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>

              {/* Order Summary */}
              <div className="space-y-6">
                <Card className="shadow-2xl border-2 border-[#1B4332]/10 rounded-3xl sticky top-24">
                  <CardContent className="p-8">
                    <h3 className="font-display text-2xl font-bold text-[#1B4332] mb-6">
                      Resumo do pedido
                    </h3>

                    <div className="space-y-4 font-body">
                      <div className="flex justify-between text-[#2D6A4F]">
                        <span>Subtotal</span>
                        <span className="font-semibold text-[#1B4332]">
                          R$ {subtotal.toFixed(2)}
                        </span>
                      </div>
                      <div className="flex justify-between text-[#2D6A4F]">
                        <span>Frete</span>
                        <span className="font-semibold text-[#1B4332]">
                          {shipping === 0 ? (
                            <span className="text-[#95D5B2] font-bold">Grátis</span>
                          ) : (
                            `R$ ${shipping.toFixed(2)}`
                          )}
                        </span>
                      </div>
                      <div className="flex justify-between text-[#2D6A4F]">
                        <span>Impostos</span>
                        <span className="font-semibold text-[#1B4332]">
                          R$ {tax.toFixed(2)}
                        </span>
                      </div>
                      {appliedDiscount > 0 && (
                        <div className="flex justify-between text-[#95D5B2] font-bold">
                          <span>Desconto</span>
                          <span>-R$ {appliedDiscount.toFixed(2)}</span>
                        </div>
                      )}
                      <Separator className="bg-[#1B4332]/20" />
                      <div className="flex justify-between text-xl pt-2">
                        <span className="font-display font-bold text-[#1B4332]">Total</span>
                        <span className="font-display font-bold text-[#52B788]">
                          R$ {total.toFixed(2)}
                        </span>
                      </div>
                    </div>

                    <div className="mt-8 space-y-4">
                      <div className="flex space-x-3">
                        <Input
                          placeholder="Cupom de desconto"
                          value={promoCode}
                          onChange={(e) => setPromoCode(e.target.value)}
                          onKeyDown={(e) => e.key === "Enter" && applyPromoCode()}
                          className="font-body border-2 border-[#1B4332]/10 rounded-xl focus:border-[#52B788]"
                        />
                        <Button
                          variant="outline"
                          onClick={applyPromoCode}
                          className="font-body border-2 border-[#1B4332] hover:bg-[#1B4332] hover:text-white rounded-xl px-6"
                        >
                          Aplicar
                        </Button>
                      </div>

                      <Button
                        className="w-full btn-primary-custom font-body text-lg rounded-full py-6"
                        disabled={cartItems.length === 0 || isProcessing}
                        onClick={handleCheckout}
                      >
                        {isProcessing ? (
                          <>
                            <Loader2 className="h-5 w-5 mr-2 animate-spin" />
                            Processando...
                          </>
                        ) : (
                          "Finalizar compra"
                        )}
                      </Button>
                    </div>
                  </CardContent>
                </Card>

                {/* Benefits */}
                <Card className="shadow-lg border-2 border-[#1B4332]/10 rounded-3xl">
                  <CardContent className="p-6">
                    <h4 className="font-display text-xl font-bold text-[#1B4332] mb-6">
                      Seus benefícios
                    </h4>

                    <div className="space-y-4">
                      <div className="flex items-start space-x-4">
                        <div className="bg-gradient-to-br from-[#52B788] to-[#40916C] rounded-2xl p-3 flex-shrink-0">
                          <Truck className="h-6 w-6 text-white" />
                        </div>
                        <div>
                          <p className="font-body font-semibold text-[#1B4332]">Frete grátis</p>
                          <p className="font-body text-sm text-[#2D6A4F]">
                            Em pedidos acima de R$ 100
                          </p>
                        </div>
                      </div>

                      <div className="flex items-start space-x-4">
                        <div className="bg-gradient-to-br from-[#95D5B2] to-[#2D6A4F] rounded-2xl p-3 flex-shrink-0">
                          <Shield className="h-6 w-6 text-white" />
                        </div>
                        <div>
                          <p className="font-body font-semibold text-[#1B4332]">
                            Devolução em 30 dias
                          </p>
                          <p className="font-body text-sm text-[#2D6A4F]">
                            Garantia de dinheiro de volta
                          </p>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}
        </main>
      </div>
    </>
  );
};

export default Cart;
