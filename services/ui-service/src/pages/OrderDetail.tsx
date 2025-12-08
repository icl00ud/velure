import { ArrowLeft, Check, CheckCircle, Clock, Loader2, Package, Truck } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import Header from "@/components/Header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { toast } from "@/hooks/use-toast";
import { type Order, orderService } from "@/services/order.service";

const orderDetailStyles = `
  @keyframes slideUp {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .page-enter {
    animation: slideUp 0.4s ease-out forwards;
  }

  .card-hover {
    transition: all 0.2s ease;
  }

  .card-hover:hover {
    transform: translateY(-2px);
    box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.1);
  }

  .item-row {
    transition: background-color 0.2s ease;
  }

  .item-row:hover {
    background-color: rgba(0, 0, 0, 0.02);
  }
`;

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
    const baseClasses = "font-medium px-4 py-2 text-sm rounded-full";

    switch (status) {
      case "CREATED":
        return (
          <Badge className={`${baseClasses} bg-amber-100 text-amber-700 border border-amber-200`}>
            <Clock className="h-4 w-4 mr-2" />
            Aguardando
          </Badge>
        );
      case "PROCESSING":
        return (
          <Badge className={`${baseClasses} bg-blue-100 text-blue-700 border border-blue-200`}>
            <Truck className="h-4 w-4 mr-2" />
            Processando
          </Badge>
        );
      case "COMPLETED":
        return (
          <Badge className={`${baseClasses} bg-emerald-100 text-emerald-700 border border-emerald-200`}>
            <CheckCircle className="h-4 w-4 mr-2" />
            Concluído
          </Badge>
        );
      default:
        return <Badge className={baseClasses}>{status}</Badge>;
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

  const isStepComplete = (status: string, step: "CREATED" | "PROCESSING" | "COMPLETED") => {
    const order = ["CREATED", "PROCESSING", "COMPLETED"];
    return order.indexOf(status) >= order.indexOf(step);
  };

  if (isLoading) {
    return (
      <>
        <style>{orderDetailStyles}</style>
        <div className="min-h-screen bg-white">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-16">
            <div className="flex flex-col justify-center items-center py-32">
              <Loader2 className="h-10 w-10 animate-spin text-slate-400" />
              <p className="text-slate-500 mt-6">Carregando pedido...</p>
            </div>
          </main>
        </div>
      </>
    );
  }

  if (!order) {
    return (
      <>
        <style>{orderDetailStyles}</style>
        <div className="min-h-screen bg-white">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-16">
            <Card className="text-center py-16 rounded-2xl border border-slate-200">
              <CardContent>
                <Package className="h-12 w-12 text-slate-300 mx-auto mb-4" />
                <h3 className="text-xl font-semibold text-slate-800 mb-2">
                  Pedido não encontrado
                </h3>
                <p className="text-slate-500 mb-8">
                  O pedido que você está procurando não existe ou foi removido.
                </p>
                <Button asChild className="bg-slate-900 hover:bg-slate-800 text-white rounded-full px-8">
                  <Link to="/orders">Voltar para meus pedidos</Link>
                </Button>
              </CardContent>
            </Card>
          </main>
        </div>
      </>
    );
  }

  const getOrderId = (order: Order) => order.id || order._id || "UNKNOWN";
  const getOrderDate = (order: Order) => {
    const dateStr = order.created_at || order.createdAt;
    if (!dateStr) return new Date();
    return new Date(dateStr);
  };

  return (
    <>
      <style>{orderDetailStyles}</style>
      <div className="min-h-screen bg-slate-50">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-10">
          {/* Header Section */}
          <div className={`mb-10 ${isVisible ? "page-enter" : "opacity-0"}`}>
            <Link
              to="/orders"
              className="inline-flex items-center text-sm text-slate-500 hover:text-slate-800 transition-colors mb-6 group"
            >
              <ArrowLeft className="h-4 w-4 mr-2 group-hover:-translate-x-1 transition-transform" />
              Voltar para meus pedidos
            </Link>

            <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
              <div>
                <h1 className="text-2xl lg:text-3xl font-bold text-slate-900 mb-1">
                  Pedido #{getOrderId(order).slice(0, 8).toUpperCase()}
                </h1>
                <p className="text-slate-500">
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

          <div className="grid lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 space-y-6">
              {/* Status Tracking Card */}
              <Card className="card-hover border border-slate-200 rounded-2xl bg-white">
                <CardHeader className="pb-4">
                  <CardTitle className="text-lg font-semibold text-slate-900">
                    Acompanhamento do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-8">
                    {/* Progress Bar - Simple */}
                    <div className="relative">
                      <div className="h-2 rounded-full bg-slate-100">
                        <div
                          style={{ width: `${getStatusProgress(order.status)}%` }}
                          className="h-full rounded-full bg-emerald-500 transition-all duration-500"
                        />
                      </div>
                    </div>

                    {/* Status Steps - Simple black and white with green check */}
                    <div className="grid grid-cols-3 gap-4">
                      {/* Step 1: Created */}
                      <div className="text-center">
                        <div className={`w-10 h-10 rounded-full flex items-center justify-center mx-auto mb-3 border-2 ${
                          isStepComplete(order.status, "CREATED")
                            ? "bg-emerald-500 border-emerald-500"
                            : "bg-white border-slate-300"
                        }`}>
                          {isStepComplete(order.status, "CREATED") ? (
                            <Check className="h-5 w-5 text-white" />
                          ) : (
                            <span className="text-sm text-slate-400">1</span>
                          )}
                        </div>
                        <p className={`text-sm font-medium ${
                          isStepComplete(order.status, "CREATED") ? "text-slate-900" : "text-slate-400"
                        }`}>
                          Criado
                        </p>
                      </div>

                      {/* Step 2: Processing */}
                      <div className="text-center">
                        <div className={`w-10 h-10 rounded-full flex items-center justify-center mx-auto mb-3 border-2 ${
                          isStepComplete(order.status, "PROCESSING")
                            ? "bg-emerald-500 border-emerald-500"
                            : "bg-white border-slate-300"
                        }`}>
                          {isStepComplete(order.status, "PROCESSING") ? (
                            <Check className="h-5 w-5 text-white" />
                          ) : (
                            <span className="text-sm text-slate-400">2</span>
                          )}
                        </div>
                        <p className={`text-sm font-medium ${
                          isStepComplete(order.status, "PROCESSING") ? "text-slate-900" : "text-slate-400"
                        }`}>
                          Processando
                        </p>
                      </div>

                      {/* Step 3: Completed */}
                      <div className="text-center">
                        <div className={`w-10 h-10 rounded-full flex items-center justify-center mx-auto mb-3 border-2 ${
                          isStepComplete(order.status, "COMPLETED")
                            ? "bg-emerald-500 border-emerald-500"
                            : "bg-white border-slate-300"
                        }`}>
                          {isStepComplete(order.status, "COMPLETED") ? (
                            <Check className="h-5 w-5 text-white" />
                          ) : (
                            <span className="text-sm text-slate-400">3</span>
                          )}
                        </div>
                        <p className={`text-sm font-medium ${
                          isStepComplete(order.status, "COMPLETED") ? "text-slate-900" : "text-slate-400"
                        }`}>
                          Concluído
                        </p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Items Card */}
              <Card className="card-hover border border-slate-200 rounded-2xl bg-white">
                <CardHeader className="pb-4">
                  <CardTitle className="text-lg font-semibold text-slate-900">
                    Itens do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-1">
                    {order.items?.map((item, index) => (
                      <div key={index}>
                        <div className="item-row flex justify-between items-center py-4 px-3 -mx-3 rounded-lg">
                          <div>
                            <p className="font-medium text-slate-900">{item.name}</p>
                            <p className="text-sm text-slate-500 mt-0.5">
                              {item.quantity} {item.quantity > 1 ? "unidades" : "unidade"} × R$ {item.price.toFixed(2)}
                            </p>
                          </div>
                          <p className="font-semibold text-slate-900">
                            R$ {(item.price * item.quantity).toFixed(2)}
                          </p>
                        </div>
                        {index < (order.items?.length || 0) - 1 && (
                          <Separator className="bg-slate-100" />
                        )}
                      </div>
                    )) || (
                      <p className="text-slate-500 text-center py-8">
                        Nenhum item encontrado
                      </p>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Summary Card */}
            <div>
              <Card className="card-hover border border-slate-200 rounded-2xl bg-white sticky top-24">
                <CardHeader className="pb-4">
                  <CardTitle className="text-lg font-semibold text-slate-900">
                    Resumo do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex justify-between py-2">
                      <span className="text-slate-500">Subtotal</span>
                      <span className="font-medium text-slate-900">
                        R$ {order.total.toFixed(2)}
                      </span>
                    </div>
                    <div className="flex justify-between py-2">
                      <span className="text-slate-500">Entrega</span>
                      <span className="font-medium text-emerald-600">Grátis</span>
                    </div>
                    <Separator className="bg-slate-100" />
                    <div className="flex justify-between items-baseline pt-2">
                      <span className="font-semibold text-slate-900">Total</span>
                      <span className="text-2xl font-bold text-slate-900">
                        R$ {order.total.toFixed(2)}
                      </span>
                    </div>
                  </div>

                  <div className="mt-8 pt-6 border-t border-slate-100">
                    <p className="text-sm text-slate-500 text-center">
                      Precisa de ajuda?{" "}
                      <Link to="/contact" className="text-slate-900 font-medium hover:underline">
                        Entre em contato
                      </Link>
                    </p>
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
