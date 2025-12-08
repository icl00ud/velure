import { ArrowLeft, CheckCircle, Clock, Loader2, Package, Sparkles, Truck } from "lucide-react";
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
  @import url('https://fonts.googleapis.com/css2?family=DM+Sans:opsz,wght@9..40,400;9..40,500;9..40,600;9..40,700&family=Space+Grotesk:wght@500;600;700&display=swap');

  @keyframes slideUp {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
  }

  @keyframes shimmer {
    0% { background-position: -200% 0; }
    100% { background-position: 200% 0; }
  }

  @keyframes pulse-ring {
    0% { transform: scale(0.9); opacity: 1; }
    50% { transform: scale(1.1); opacity: 0.5; }
    100% { transform: scale(0.9); opacity: 1; }
  }

  @keyframes bounce-subtle {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-3px); }
  }

  @keyframes glow {
    0%, 100% { box-shadow: 0 0 20px rgba(16, 185, 129, 0.3); }
    50% { box-shadow: 0 0 30px rgba(16, 185, 129, 0.5); }
  }

  .page-enter {
    animation: slideUp 0.5s ease-out forwards;
  }

  .card-float {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .card-float:hover {
    transform: translateY(-4px);
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.12);
  }

  .progress-bar {
    background: linear-gradient(90deg, #10B981 0%, #34D399 50%, #6EE7B7 100%);
    background-size: 200% 100%;
    animation: shimmer 2s infinite;
  }

  .status-active-ring {
    animation: pulse-ring 2s ease-in-out infinite;
  }

  .badge-bounce {
    animation: bounce-subtle 2s ease-in-out infinite;
  }

  .success-glow {
    animation: glow 2s ease-in-out infinite;
  }

  .item-row {
    transition: all 0.2s ease;
  }

  .item-row:hover {
    background: linear-gradient(90deg, rgba(16, 185, 129, 0.05) 0%, transparent 100%);
    transform: translateX(4px);
  }

  .step-connector {
    background: linear-gradient(90deg, currentColor 50%, transparent 50%);
    background-size: 8px 2px;
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
    const baseClasses = "font-semibold px-4 py-2 text-sm tracking-wide rounded-full badge-bounce shadow-lg";

    switch (status) {
      case "CREATED":
        return (
          <Badge className={`${baseClasses} bg-gradient-to-r from-amber-400 to-orange-400 text-white border-0`}>
            <Clock className="h-4 w-4 mr-2" />
            Aguardando
          </Badge>
        );
      case "PROCESSING":
        return (
          <Badge className={`${baseClasses} bg-gradient-to-r from-blue-500 to-indigo-500 text-white border-0`}>
            <Truck className="h-4 w-4 mr-2" />
            Processando
          </Badge>
        );
      case "COMPLETED":
        return (
          <Badge className={`${baseClasses} bg-gradient-to-r from-emerald-500 to-teal-500 text-white border-0 success-glow`}>
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

  const isStepActive = (status: string, step: "CREATED" | "PROCESSING" | "COMPLETED") => {
    const order = ["CREATED", "PROCESSING", "COMPLETED"];
    return order.indexOf(status) >= order.indexOf(step);
  };

  const isCurrentStep = (status: string, step: string) => status === step;

  if (isLoading) {
    return (
      <>
        <style>{orderDetailStyles}</style>
        <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-emerald-50/30">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-16">
            <div className="flex flex-col justify-center items-center py-32">
              <div className="relative">
                <div className="absolute inset-0 bg-emerald-400/30 blur-3xl rounded-full scale-150" />
                <div className="relative bg-white rounded-full p-6 shadow-xl">
                  <Loader2 className="h-12 w-12 animate-spin text-emerald-500" />
                </div>
              </div>
              <p className="text-lg text-slate-600 mt-8 font-medium tracking-wide">
                Carregando pedido...
              </p>
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
        <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-emerald-50/30">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-16">
            <Card className="text-center py-20 rounded-3xl border-0 shadow-2xl bg-white page-enter">
              <CardContent>
                <div className="relative inline-block mb-8">
                  <div className="absolute inset-0 bg-slate-200 blur-2xl scale-150" />
                  <div className="relative bg-gradient-to-br from-slate-100 to-slate-200 rounded-3xl p-8">
                    <Package className="h-16 w-16 text-slate-400" />
                  </div>
                </div>
                <h3 className="text-3xl font-bold text-slate-800 mb-4">
                  Pedido não encontrado
                </h3>
                <p className="text-slate-500 mb-10 max-w-md mx-auto text-lg">
                  O pedido que você está procurando não existe ou foi removido.
                </p>
                <Button
                  asChild
                  className="bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 text-white font-semibold px-10 py-6 rounded-full text-base shadow-xl shadow-emerald-500/25 transition-all duration-300 hover:shadow-2xl hover:shadow-emerald-500/30"
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

  const getOrderId = (order: Order) => order.id || order._id || "UNKNOWN";
  const getOrderDate = (order: Order) => {
    const dateStr = order.created_at || order.createdAt;
    if (!dateStr) return new Date();
    return new Date(dateStr);
  };

  return (
    <>
      <style>{orderDetailStyles}</style>
      <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-emerald-50/30">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-10">
          {/* Header Section */}
          <div className={`mb-10 ${isVisible ? "page-enter" : "opacity-0"}`}>
            <Link
              to="/orders"
              className="inline-flex items-center text-sm text-slate-500 hover:text-emerald-600 transition-colors mb-6 group font-medium"
            >
              <ArrowLeft className="h-4 w-4 mr-2 group-hover:-translate-x-1 transition-transform" />
              Voltar para meus pedidos
            </Link>

            <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-6">
              <div>
                <div className="flex items-center gap-3 mb-2">
                  <div className="bg-gradient-to-r from-emerald-500 to-teal-500 rounded-xl p-2">
                    <Package className="h-5 w-5 text-white" />
                  </div>
                  <h1 className="text-3xl lg:text-4xl font-bold text-slate-800 tracking-tight">
                    Pedido #{getOrderId(order).slice(0, 8).toUpperCase()}
                  </h1>
                </div>
                <p className="text-slate-500 text-lg">
                  Realizado em{" "}
                  <span className="text-slate-700 font-medium">
                    {getOrderDate(order).toLocaleDateString("pt-BR", {
                      day: "2-digit",
                      month: "long",
                      year: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </span>
                </p>
              </div>
              {getStatusBadge(order.status)}
            </div>
          </div>

          <div className="grid lg:grid-cols-3 gap-8">
            <div className="lg:col-span-2 space-y-8">
              {/* Status Tracking Card */}
              <Card className="card-float shadow-xl border-0 rounded-3xl bg-white overflow-hidden" style={{ animationDelay: "0.1s" }}>
                <div className="h-1.5 bg-gradient-to-r from-emerald-400 via-teal-400 to-cyan-400" />
                <CardHeader className="pb-2 pt-6">
                  <CardTitle className="text-xl font-bold text-slate-800 flex items-center gap-3">
                    <Sparkles className="h-5 w-5 text-emerald-500" />
                    Acompanhamento do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-4 pb-8">
                  <div className="space-y-8">
                    {/* Progress Bar */}
                    <div className="relative px-2">
                      <div className="overflow-hidden h-3 rounded-full bg-slate-100">
                        <div
                          style={{ width: `${getStatusProgress(order.status)}%` }}
                          className="h-full rounded-full progress-bar transition-all duration-700 ease-out"
                        />
                      </div>
                      <div className="flex justify-between mt-2 text-xs text-slate-400 font-medium">
                        <span>Início</span>
                        <span>Concluído</span>
                      </div>
                    </div>

                    {/* Status Steps */}
                    <div className="grid grid-cols-3 gap-4">
                      {/* Step 1: Created */}
                      <div
                        className={`relative p-6 rounded-2xl border-2 transition-all duration-300 ${
                          isStepActive(order.status, "CREATED")
                            ? "bg-gradient-to-br from-amber-50 to-orange-50 border-amber-300"
                            : "bg-slate-50 border-slate-200"
                        } ${isCurrentStep(order.status, "CREATED") ? "ring-4 ring-amber-200" : ""}`}
                      >
                        <div className={`w-14 h-14 rounded-2xl flex items-center justify-center mx-auto mb-4 ${
                          isStepActive(order.status, "CREATED")
                            ? "bg-gradient-to-br from-amber-400 to-orange-400 shadow-lg shadow-amber-400/30"
                            : "bg-slate-200"
                        } ${isCurrentStep(order.status, "CREATED") ? "status-active-ring" : ""}`}>
                          <Clock className={`h-7 w-7 ${
                            isStepActive(order.status, "CREATED") ? "text-white" : "text-slate-400"
                          }`} />
                        </div>
                        <p className={`text-sm font-bold text-center ${
                          isStepActive(order.status, "CREATED") ? "text-amber-700" : "text-slate-400"
                        }`}>
                          Criado
                        </p>
                        <p className={`text-xs text-center mt-1 ${
                          isStepActive(order.status, "CREATED") ? "text-amber-600/70" : "text-slate-400"
                        }`}>
                          Pedido recebido
                        </p>
                      </div>

                      {/* Step 2: Processing */}
                      <div
                        className={`relative p-6 rounded-2xl border-2 transition-all duration-300 ${
                          isStepActive(order.status, "PROCESSING")
                            ? "bg-gradient-to-br from-blue-50 to-indigo-50 border-blue-300"
                            : "bg-slate-50 border-slate-200"
                        } ${isCurrentStep(order.status, "PROCESSING") ? "ring-4 ring-blue-200" : ""}`}
                      >
                        <div className={`w-14 h-14 rounded-2xl flex items-center justify-center mx-auto mb-4 ${
                          isStepActive(order.status, "PROCESSING")
                            ? "bg-gradient-to-br from-blue-500 to-indigo-500 shadow-lg shadow-blue-500/30"
                            : "bg-slate-200"
                        } ${isCurrentStep(order.status, "PROCESSING") ? "status-active-ring" : ""}`}>
                          <Truck className={`h-7 w-7 ${
                            isStepActive(order.status, "PROCESSING") ? "text-white" : "text-slate-400"
                          }`} />
                        </div>
                        <p className={`text-sm font-bold text-center ${
                          isStepActive(order.status, "PROCESSING") ? "text-blue-700" : "text-slate-400"
                        }`}>
                          Processando
                        </p>
                        <p className={`text-xs text-center mt-1 ${
                          isStepActive(order.status, "PROCESSING") ? "text-blue-600/70" : "text-slate-400"
                        }`}>
                          Em preparação
                        </p>
                      </div>

                      {/* Step 3: Completed */}
                      <div
                        className={`relative p-6 rounded-2xl border-2 transition-all duration-300 ${
                          isStepActive(order.status, "COMPLETED")
                            ? "bg-gradient-to-br from-emerald-50 to-teal-50 border-emerald-300"
                            : "bg-slate-50 border-slate-200"
                        } ${isCurrentStep(order.status, "COMPLETED") ? "ring-4 ring-emerald-200" : ""}`}
                      >
                        <div className={`w-14 h-14 rounded-2xl flex items-center justify-center mx-auto mb-4 ${
                          isStepActive(order.status, "COMPLETED")
                            ? "bg-gradient-to-br from-emerald-500 to-teal-500 shadow-lg shadow-emerald-500/30"
                            : "bg-slate-200"
                        }`}>
                          <CheckCircle className={`h-7 w-7 ${
                            isStepActive(order.status, "COMPLETED") ? "text-white" : "text-slate-400"
                          }`} />
                        </div>
                        <p className={`text-sm font-bold text-center ${
                          isStepActive(order.status, "COMPLETED") ? "text-emerald-700" : "text-slate-400"
                        }`}>
                          Concluído
                        </p>
                        <p className={`text-xs text-center mt-1 ${
                          isStepActive(order.status, "COMPLETED") ? "text-emerald-600/70" : "text-slate-400"
                        }`}>
                          Entregue
                        </p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Items Card */}
              <Card className="card-float shadow-xl border-0 rounded-3xl bg-white" style={{ animationDelay: "0.2s" }}>
                <CardHeader className="pb-2 pt-6">
                  <CardTitle className="text-xl font-bold text-slate-800">
                    Itens do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-4">
                  <div className="space-y-1">
                    {order.items?.map((item, index) => (
                      <div key={index}>
                        <div className="item-row flex justify-between items-center py-5 px-4 -mx-4 rounded-xl">
                          <div className="flex items-center gap-4">
                            <div className="w-12 h-12 bg-gradient-to-br from-emerald-100 to-teal-100 rounded-xl flex items-center justify-center">
                              <Package className="h-6 w-6 text-emerald-600" />
                            </div>
                            <div>
                              <p className="text-base font-semibold text-slate-800">
                                {item.name}
                              </p>
                              <p className="text-sm text-slate-500 mt-0.5">
                                {item.quantity} {item.quantity > 1 ? "unidades" : "unidade"} × R$ {item.price.toFixed(2)}
                              </p>
                            </div>
                          </div>
                          <p className="text-lg font-bold text-emerald-600">
                            R$ {(item.price * item.quantity).toFixed(2)}
                          </p>
                        </div>
                        {index < (order.items?.length || 0) - 1 && (
                          <Separator className="bg-slate-100" />
                        )}
                      </div>
                    )) || (
                      <div className="text-center py-12">
                        <Package className="h-12 w-12 text-slate-300 mx-auto mb-4" />
                        <p className="text-slate-500">
                          Nenhum item encontrado
                        </p>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Summary Card */}
            <div>
              <Card className="card-float shadow-xl border-0 rounded-3xl bg-white sticky top-24 overflow-hidden" style={{ animationDelay: "0.3s" }}>
                <div className="h-2 bg-gradient-to-r from-emerald-400 via-teal-400 to-cyan-400" />
                <CardHeader className="pb-2 pt-6">
                  <CardTitle className="text-xl font-bold text-slate-800">
                    Resumo do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-4">
                  <div className="space-y-4">
                    <div className="flex justify-between py-3">
                      <span className="text-slate-500">Subtotal</span>
                      <span className="font-semibold text-slate-700">
                        R$ {order.total.toFixed(2)}
                      </span>
                    </div>
                    <div className="flex justify-between py-3">
                      <span className="text-slate-500">Entrega</span>
                      <span className="font-semibold text-emerald-500 bg-emerald-50 px-3 py-1 rounded-full text-sm">
                        Grátis
                      </span>
                    </div>
                    <Separator className="bg-slate-100" />
                    <div className="flex justify-between items-baseline pt-4 pb-2">
                      <span className="font-bold text-slate-800 text-lg">Total</span>
                      <div className="text-right">
                        <span className="text-3xl font-bold bg-gradient-to-r from-emerald-500 to-teal-500 bg-clip-text text-transparent">
                          R$ {order.total.toFixed(2)}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Help Section */}
                  <div className="mt-8 p-4 bg-gradient-to-br from-slate-50 to-slate-100 rounded-2xl">
                    <p className="text-sm text-slate-600 text-center">
                      Precisa de ajuda?{" "}
                      <Link to="/contact" className="text-emerald-600 font-semibold hover:text-emerald-700 underline underline-offset-2">
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
