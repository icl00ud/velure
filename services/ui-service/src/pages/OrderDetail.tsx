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
    const eventSource = orderService.createOrderStatusStream(id);

    eventSource.onmessage = (event) => {
      try {
        const updatedOrder = JSON.parse(event.data);
        setOrder(updatedOrder);
      } catch (error) {
        console.error("Failed to parse SSE data:", error);
      }
    };

    eventSource.onerror = (error) => {
      console.error("SSE connection error:", error);
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
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
        <div className="min-h-screen bg-[#FAF7F2]">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-12">
            <div className="flex flex-col justify-center items-center py-32">
              <Loader2 className="h-16 w-16 animate-spin text-[#D97757] mb-4" />
              <p className="font-body text-lg text-[#5A6751]">Carregando pedido...</p>
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
        <div className="min-h-screen bg-[#FAF7F2]">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-12">
            <Card className="text-center py-20 rounded-3xl border-2 border-[#2D3319]/10 shadow-2xl">
              <CardContent>
                <div className="relative inline-block mb-8">
                  <div className="absolute inset-0 bg-[#D97757]/20 blur-3xl" />
                  <div className="relative bg-gradient-to-br from-[#D97757]/10 to-[#8B9A7E]/10 rounded-full p-8">
                    <Package className="h-20 w-20 text-[#D97757]" />
                  </div>
                </div>
                <h3 className="font-display text-3xl font-bold text-[#2D3319] mb-4">
                  Pedido não encontrado
                </h3>
                <p className="font-body text-lg text-[#5A6751] mb-8 max-w-md mx-auto">
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
      <div className="min-h-screen bg-[#FAF7F2]">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-12">
          <div className={`mb-12 ${isVisible ? 'hero-enter active' : 'hero-enter'}`}>
            <Link
              to="/orders"
              className="inline-flex items-center font-body text-[#5A6751] hover:text-[#D97757] transition-colors mb-6 group"
            >
              <ArrowLeft className="h-5 w-5 mr-2 group-hover:-translate-x-1 transition-transform" />
              Voltar para meus pedidos
            </Link>
            <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
              <div>
                <h1 className="font-display text-4xl lg:text-5xl font-bold text-[#2D3319] mb-4">
                  Pedido #{getOrderId(order).slice(0, 8)}
                </h1>
                <p className="font-body text-lg text-[#5A6751]">
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
              <Card className="shadow-2xl border-2 border-[#2D3319]/10 rounded-3xl observe-animation">
                <CardHeader>
                  <CardTitle className="font-display text-2xl font-bold text-[#2D3319]">
                    Status do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-6">
                    <div className="relative pt-1">
                      <div className="overflow-hidden h-3 text-xs flex rounded-full bg-gray-200">
                        <div
                          style={{ width: `${getStatusProgress(order.status)}%` }}
                          className="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-gradient-to-r from-[#D97757] to-[#C56647] transition-all duration-500"
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-3 gap-4 text-center">
                      <div
                        className={`p-5 rounded-2xl border-2 ${
                          order.status === "CREATED" ||
                          order.status === "PROCESSING" ||
                          order.status === "COMPLETED"
                            ? "bg-[#D97757]/10 border-[#D97757]"
                            : "bg-gray-100 border-gray-200"
                        }`}
                      >
                        <Clock
                          className={`h-8 w-8 mx-auto mb-3 ${
                            order.status === "CREATED" ||
                            order.status === "PROCESSING" ||
                            order.status === "COMPLETED"
                              ? "text-[#D97757]"
                              : "text-gray-400"
                          }`}
                        />
                        <p className="font-body text-sm font-semibold">Criado</p>
                      </div>
                      <div
                        className={`p-5 rounded-2xl border-2 ${
                          order.status === "PROCESSING" || order.status === "COMPLETED"
                            ? "bg-[#8B9A7E]/10 border-[#8B9A7E]"
                            : "bg-gray-100 border-gray-200"
                        }`}
                      >
                        <Package
                          className={`h-8 w-8 mx-auto mb-3 ${
                            order.status === "PROCESSING" || order.status === "COMPLETED"
                              ? "text-[#8B9A7E]"
                              : "text-gray-400"
                          }`}
                        />
                        <p className="font-body text-sm font-semibold">Processando</p>
                      </div>
                      <div
                        className={`p-5 rounded-2xl border-2 ${
                          order.status === "COMPLETED"
                            ? "bg-[#F4C430]/10 border-[#F4C430]"
                            : "bg-gray-100 border-gray-200"
                        }`}
                      >
                        <CheckCircle
                          className={`h-8 w-8 mx-auto mb-3 ${
                            order.status === "COMPLETED" ? "text-[#F4C430]" : "text-gray-400"
                          }`}
                        />
                        <p className="font-body text-sm font-semibold">Concluído</p>
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

              <Card className="shadow-2xl border-2 border-[#2D3319]/10 rounded-3xl observe-animation" style={{ animationDelay: '0.1s' }}>
                <CardHeader>
                  <CardTitle className="font-display text-2xl font-bold text-[#2D3319]">
                    Itens do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-6">
                    {order.items?.map((item, index) => (
                      <div key={index}>
                        <div className="flex justify-between items-start py-4">
                          <div>
                            <p className="font-display text-lg font-bold text-[#2D3319]">
                              {item.name}
                            </p>
                            <p className="font-body text-sm text-[#5A6751] mt-1">
                              Quantidade: <span className="font-semibold">{item.quantity}</span>
                            </p>
                          </div>
                          <div className="text-right">
                            <p className="font-display text-xl font-bold text-[#D97757]">
                              R$ {(item.price * item.quantity).toFixed(2)}
                            </p>
                            <p className="font-body text-sm text-[#5A6751]">
                              R$ {item.price.toFixed(2)} cada
                            </p>
                          </div>
                        </div>
                        {index < (order.items?.length || 0) - 1 && (
                          <Separator className="bg-[#2D3319]/20" />
                        )}
                      </div>
                    )) || (
                      <p className="font-body text-[#5A6751] text-center py-8">
                        Nenhum item encontrado
                      </p>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>

            <div>
              <Card className="shadow-2xl border-2 border-[#2D3319]/10 rounded-3xl sticky top-24 observe-animation" style={{ animationDelay: '0.2s' }}>
                <CardHeader>
                  <CardTitle className="font-display text-2xl font-bold text-[#2D3319]">
                    Resumo do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex justify-between py-3 font-body">
                      <span className="text-[#5A6751]">Subtotal</span>
                      <span className="font-semibold text-[#2D3319]">
                        R$ {order.total.toFixed(2)}
                      </span>
                    </div>
                    <Separator className="bg-[#2D3319]/20" />
                    <div className="flex justify-between text-xl pt-2 font-display">
                      <span className="font-bold text-[#2D3319]">Total</span>
                      <span className="font-bold text-[#D97757]">
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
