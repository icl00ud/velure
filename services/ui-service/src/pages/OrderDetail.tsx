import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { ArrowLeft, Clock, Package, CheckCircle, Loader2 } from "lucide-react";
import Header from "@/components/Header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { orderService, type Order } from "@/services/order.service";
import { toast } from "@/hooks/use-toast";
import { designSystemStyles } from "@/styles/design-system";

const OrderDetail = () => {
  const { id } = useParams<{ id: string }>();
  const [order, setOrder] = useState<Order | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    setIsVisible(true);
  }, []);

  useEffect(() => {
    if (!id) return;

    loadOrder();

    const cleanup = orderService.createOrderStatusStream(
      id,
      (updatedOrder) => setOrder(updatedOrder),
      (error) => console.error("SSE connection error:", error)
    );

    return cleanup;
  }, [id]);

  const loadOrder = async () => {
    if (!id) return;

    setIsLoading(true);
    try {
      const result = await orderService.getUserOrderById(id);
      setOrder(result);
    } catch (error) {
      toast({
        title: "Erro ao carregar pedido",
        description: error instanceof Error ? error.message : "Tente novamente mais tarde",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "CREATED":
        return (
          <Badge className="bg-yellow-100 text-yellow-700 border-2 border-yellow-300 font-body font-semibold px-5 py-2 text-base">
            <Clock className="h-5 w-5 mr-2" />
            Criado
          </Badge>
        );
      case "PROCESSING":
        return (
          <Badge className="bg-blue-100 text-blue-700 border-2 border-blue-300 font-body font-semibold px-5 py-2 text-base">
            <Package className="h-5 w-5 mr-2" />
            Processando
          </Badge>
        );
      case "COMPLETED":
        return (
          <Badge className="bg-green-100 text-green-700 border-2 border-green-300 font-body font-semibold px-5 py-2 text-base">
            <CheckCircle className="h-5 w-5 mr-2" />
            Concluído
          </Badge>
        );
      default:
        return <Badge className="font-body font-semibold px-5 py-2 text-base">{status}</Badge>;
    }
  };

  const getStatusProgress = (status: string) => {
    switch (status) {
      case "CREATED":
        return 33;
      case "PROCESSING":
        return 66;
      case "COMPLETED":
        return 100;
      default:
        return 0;
    }
  };

  if (isLoading) {
    return (
      <>
        <style>{designSystemStyles}</style>
        <div className="min-h-screen bg-[#F8FAF5]">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-12">
            <div className="flex flex-col justify-center items-center py-32">
              <Loader2 className="h-16 w-16 animate-spin text-[#52B788] mb-4" />
              <p className="font-body text-lg text-[#2D6A4F]">Carregando pedido...</p>
            </div>
          </main>
        </div>
      </>
    );
  }

  if (!order) {
    return (
      <>
        <style>{designSystemStyles}</style>
        <div className="min-h-screen bg-[#F8FAF5]">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-12">
            <Card className="text-center py-20 rounded-3xl border-2 border-[#1B4332]/10 shadow-2xl">
              <CardContent>
                <div className="relative inline-block mb-8">
                  <div className="absolute inset-0 bg-[#52B788]/20 blur-3xl" />
                  <div className="relative bg-gradient-to-br from-[#52B788]/10 to-[#95D5B2]/10 rounded-full p-8">
                    <Package className="h-20 w-20 text-[#52B788]" />
                  </div>
                </div>
                <h3 className="font-display text-3xl font-bold text-[#1B4332] mb-4">
                  Pedido não encontrado
                </h3>
                <p className="font-body text-lg text-[#2D6A4F] mb-8 max-w-md mx-auto">
                  O pedido que você está procurando não existe ou foi removido.
                </p>
                <Button
                  asChild
                  className="btn-primary-custom font-body px-10 py-6 rounded-full text-lg"
                >
                  <Link to="/orders">Voltar para meus pedidos</Link>
                </Button>
              </CardContent>
            </Card>
          </main>
        </div>
      </>
    );
  }

  const getOrderId = (order: Order) => order.id || order._id || 'UNKNOWN';
  const getOrderDate = (order: Order) => {
    const dateStr = order.created_at || order.createdAt;
    if (!dateStr) return new Date();
    return new Date(dateStr);
  };

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-[#F8FAF5]">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-12">
          <div className={`mb-12 ${isVisible ? 'hero-enter active' : 'hero-enter'}`}>
            <Link
              to="/orders"
              className="inline-flex items-center font-body text-[#2D6A4F] hover:text-[#52B788] transition-colors mb-6 group"
            >
              <ArrowLeft className="h-5 w-5 mr-2 group-hover:-translate-x-1 transition-transform" />
              Voltar para meus pedidos
            </Link>
            <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
              <div>
                <h1 className="font-display text-4xl lg:text-5xl font-bold text-[#1B4332] mb-4">
                  Pedido #{getOrderId(order).slice(0, 8)}
                </h1>
                <p className="font-body text-lg text-[#2D6A4F]">
                  Realizado em{" "}
                  {getOrderDate(order).toLocaleDateString("pt-BR", {
                    day: "2-digit",
                    month: "long",
                    year: "numeric",
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </p>
              </div>
              {getStatusBadge(order.status)}
            </div>
          </div>

          <div className="grid lg:grid-cols-3 gap-8">
            <div className="lg:col-span-2 space-y-8">
              <Card className="shadow-2xl border-2 border-[#1B4332]/10 rounded-3xl">
                <CardHeader>
                  <CardTitle className="font-display text-2xl font-bold text-[#1B4332]">
                    Status do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-6">
                    <div className="relative pt-1">
                      <div className="overflow-hidden h-3 text-xs flex rounded-full bg-gray-200">
                        <div
                          style={{ width: `${getStatusProgress(order.status)}%` }}
                          className="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-gradient-to-r from-green-500 to-green-600 transition-all duration-500"
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-3 gap-4 text-center">
                      <div
                        className={`p-5 rounded-2xl border-2 ${
                          ["CREATED", "PROCESSING", "COMPLETED"].includes(order.status)
                            ? "bg-green-50 border-green-500"
                            : "bg-gray-100 border-gray-200"
                        }`}
                      >
                        <Clock
                          className={`h-8 w-8 mx-auto mb-3 ${
                            ["CREATED", "PROCESSING", "COMPLETED"].includes(order.status)
                              ? "text-green-600"
                              : "text-gray-400"
                          }`}
                        />
                        <p className={`font-body text-sm font-semibold ${
                          ["CREATED", "PROCESSING", "COMPLETED"].includes(order.status)
                            ? "text-green-700"
                            : "text-gray-500"
                        }`}>Criado</p>
                      </div>
                      <div
                        className={`p-5 rounded-2xl border-2 ${
                          ["PROCESSING", "COMPLETED"].includes(order.status)
                            ? "bg-green-50 border-green-500"
                            : "bg-gray-100 border-gray-200"
                        }`}
                      >
                        <Package
                          className={`h-8 w-8 mx-auto mb-3 ${
                            ["PROCESSING", "COMPLETED"].includes(order.status)
                              ? "text-green-600"
                              : "text-gray-400"
                          }`}
                        />
                        <p className={`font-body text-sm font-semibold ${
                          ["PROCESSING", "COMPLETED"].includes(order.status)
                            ? "text-green-700"
                            : "text-gray-500"
                        }`}>Processando</p>
                      </div>
                      <div
                        className={`p-5 rounded-2xl border-2 ${
                          order.status === "COMPLETED"
                            ? "bg-green-50 border-green-500"
                            : "bg-gray-100 border-gray-200"
                        }`}
                      >
                        <CheckCircle
                          className={`h-8 w-8 mx-auto mb-3 ${
                            order.status === "COMPLETED" ? "text-green-600" : "text-gray-400"
                          }`}
                        />
                        <p className={`font-body text-sm font-semibold ${
                          order.status === "COMPLETED"
                            ? "text-green-700"
                            : "text-gray-500"
                        }`}>Concluído</p>
                      </div>
                    </div>

                    <div className="bg-blue-50 border-2 border-blue-200 rounded-2xl p-5">
                      <p className="font-body text-sm text-blue-800 flex items-start gap-3">
                        <span className="text-2xl">ℹ️</span>
                        <span>
                          O status do seu pedido é atualizado automaticamente em tempo real.
                        </span>
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card className="shadow-2xl border-2 border-[#1B4332]/10 rounded-3xl">
                <CardHeader>
                  <CardTitle className="font-display text-2xl font-bold text-[#1B4332]">
                    Itens do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-6">
                    {order.items?.map((item, index) => (
                      <div key={index}>
                        <div className="flex justify-between items-start py-4">
                          <div>
                            <p className="font-display text-lg font-bold text-[#1B4332]">
                              {item.name}
                            </p>
                            <p className="font-body text-sm text-[#2D6A4F] mt-1">
                              Quantidade: <span className="font-semibold">{item.quantity}</span>
                            </p>
                          </div>
                          <div className="text-right">
                            <p className="font-display text-xl font-bold text-[#52B788]">
                              R$ {(item.price * item.quantity).toFixed(2)}
                            </p>
                            <p className="font-body text-sm text-[#2D6A4F]">
                              R$ {item.price.toFixed(2)} cada
                            </p>
                          </div>
                        </div>
                        {index < (order.items?.length || 0) - 1 && (
                          <Separator className="bg-[#1B4332]/20" />
                        )}
                      </div>
                    )) || (
                      <p className="font-body text-[#2D6A4F] text-center py-8">
                        Nenhum item encontrado
                      </p>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>

            <div>
              <Card className="shadow-2xl border-2 border-[#1B4332]/10 rounded-3xl sticky top-24">
                <CardHeader>
                  <CardTitle className="font-display text-2xl font-bold text-[#1B4332]">
                    Resumo do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex justify-between py-3 font-body">
                      <span className="text-[#2D6A4F]">Subtotal</span>
                      <span className="font-semibold text-[#1B4332]">
                        R$ {order.total.toFixed(2)}
                      </span>
                    </div>
                    <Separator className="bg-[#1B4332]/20" />
                    <div className="flex justify-between text-xl pt-2 font-display">
                      <span className="font-bold text-[#1B4332]">Total</span>
                      <span className="font-bold text-[#52B788]">
                        R$ {order.total.toFixed(2)}
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </main>
      </div>
    </>
  );
};

export default OrderDetail;
