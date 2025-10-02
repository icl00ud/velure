import { ArrowLeft, Loader2, Minus, Plus, Shield, ShoppingBag, Trash2, Truck } from "lucide-react";
import { useState } from "react";
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

const Cart = () => {
  const navigate = useNavigate();
  const { cartItems, updateQuantity, removeFromCart, clearCart, totalPrice } = useCart();
  const [promoCode, setPromoCode] = useState("");
  const [appliedDiscount, setAppliedDiscount] = useState(0);
  const [isProcessing, setIsProcessing] = useState(false);

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
    setIsProcessing(true);
    try {
      const order = await orderService.createOrder(cartItems);

      toast({
        title: "Pedido criado com sucesso!",
        description: `Seu pedido #${order.order_id} foi criado. Total: R$ ${(order.total / 100).toFixed(2)}`,
      });

      // Limpar carrinho após criar pedido
      clearCart();

      // Redirecionar para página de sucesso (você pode criar essa página depois)
      navigate("/");
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
  const tax = subtotal * 0.05; // 5% de impostos
  const total = subtotal + shipping + tax - appliedDiscount;

  return (
    <div className="min-h-screen bg-background">
      <Header />

      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <Link
            to="/"
            className="inline-flex items-center text-muted-foreground hover:text-primary transition-colors mb-4"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Continuar comprando
          </Link>
          <h1 className="text-4xl font-bold text-foreground">Carrinho de compras</h1>
          <p className="text-muted-foreground mt-2">
            {cartItems.length} {cartItems.length === 1 ? "item" : "itens"} no seu carrinho
          </p>
        </div>

        {cartItems.length === 0 ? (
          <Card className="text-center py-12">
            <CardContent>
              <ShoppingBag className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-xl font-semibold text-foreground mb-2">
                Seu carrinho está vazio
              </h3>
              <p className="text-muted-foreground mb-6">
                Parece que você ainda não adicionou nenhum item ao seu carrinho.
              </p>
              <Button
                asChild
                className="bg-gradient-primary hover:opacity-90 text-primary-foreground"
              >
                <Link to="/products">Começar a comprar</Link>
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid lg:grid-cols-3 gap-8">
            {/* Cart Items */}
            <div className="lg:col-span-2 space-y-4">
              {cartItems.map((item) => (
                <Card key={item.product._id} className="shadow-soft">
                  <CardContent className="p-6">
                    <div className="flex gap-4">
                      <div className="w-24 h-24 rounded-lg overflow-hidden flex-shrink-0">
                        <ProductImageWithFallback
                          images={item.product.images || []}
                          alt={item.product.name}
                          className="w-full h-full object-cover"
                          fallbackIcon="🐕"
                        />
                      </div>

                      <div className="flex-1">
                        <div className="flex justify-between items-start mb-2">
                          <div>
                            <Link
                              to={`/product/${item.product._id}`}
                              className="font-semibold text-foreground hover:text-primary transition-colors"
                            >
                              {item.product.name}
                            </Link>
                            <p className="text-sm text-muted-foreground">
                              {item.product.brand || "Marca"} •{" "}
                              {item.product.category || "Categoria"}
                            </p>
                            {!item.product.disponibility && (
                              <p className="text-sm text-destructive mt-1">Sem estoque</p>
                            )}
                          </div>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleRemoveItem(item.product._id, item.product.name)}
                            className="text-muted-foreground hover:text-destructive"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>

                        <div className="flex items-center justify-between">
                          <div className="flex items-center space-x-2">
                            <Button
                              variant="outline"
                              size="icon"
                              className="h-8 w-8"
                              onClick={() => updateQuantity(item.product._id, item.quantity - 1)}
                              disabled={item.quantity <= 1}
                            >
                              <Minus className="h-3 w-3" />
                            </Button>
                            <span className="w-8 text-center font-medium">{item.quantity}</span>
                            <Button
                              variant="outline"
                              size="icon"
                              className="h-8 w-8"
                              onClick={() => updateQuantity(item.product._id, item.quantity + 1)}
                            >
                              <Plus className="h-3 w-3" />
                            </Button>
                          </div>

                          <div className="text-right">
                            <div className="font-semibold text-primary">
                              R$ {(item.product.price * item.quantity).toFixed(2)}
                            </div>
                            <div className="text-sm text-muted-foreground">
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
              <Card className="shadow-soft">
                <CardContent className="p-6">
                  <h3 className="font-semibold text-foreground mb-4">Resumo do pedido</h3>

                  <div className="space-y-3">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Subtotal</span>
                      <span className="font-medium">R$ {subtotal.toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Frete</span>
                      <span className="font-medium">
                        {shipping === 0 ? "Grátis" : `R$ ${shipping.toFixed(2)}`}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Impostos</span>
                      <span className="font-medium">R$ {tax.toFixed(2)}</span>
                    </div>
                    {appliedDiscount > 0 && (
                      <div className="flex justify-between text-accent">
                        <span>Desconto</span>
                        <span>-R$ {appliedDiscount.toFixed(2)}</span>
                      </div>
                    )}
                    <Separator />
                    <div className="flex justify-between text-lg font-semibold">
                      <span>Total</span>
                      <span className="text-primary">R$ {total.toFixed(2)}</span>
                    </div>
                  </div>

                  <div className="mt-6 space-y-4">
                    <div className="flex space-x-2">
                      <Input
                        placeholder="Cupom de desconto"
                        value={promoCode}
                        onChange={(e) => setPromoCode(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && applyPromoCode()}
                      />
                      <Button variant="outline" onClick={applyPromoCode}>
                        Aplicar
                      </Button>
                    </div>

                    <Button
                      className="w-full bg-gradient-primary hover:opacity-90 text-primary-foreground"
                      disabled={cartItems.length === 0 || isProcessing}
                      onClick={handleCheckout}
                    >
                      {isProcessing ? (
                        <>
                          <Loader2 className="h-4 w-4 mr-2 animate-spin" />
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
              <Card className="shadow-soft">
                <CardContent className="p-6">
                  <h4 className="font-semibold text-foreground mb-4">Seus benefícios</h4>

                  <div className="space-y-3">
                    <div className="flex items-center space-x-3">
                      <div className="bg-primary/10 rounded-full p-2">
                        <Truck className="h-4 w-4 text-primary" />
                      </div>
                      <div>
                        <p className="text-sm font-medium">Frete grátis</p>
                        <p className="text-xs text-muted-foreground">Em pedidos acima de R$ 100</p>
                      </div>
                    </div>

                    <div className="flex items-center space-x-3">
                      <div className="bg-secondary/10 rounded-full p-2">
                        <Shield className="h-4 w-4 text-secondary" />
                      </div>
                      <div>
                        <p className="text-sm font-medium">Devolução em 30 dias</p>
                        <p className="text-xs text-muted-foreground">
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
  );
};

export default Cart;
