import { ArrowLeft, CheckCircle, Clock, Loader2, Package, Sparkles } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import Header from "@/components/Header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { toast } from "@/hooks/use-toast";
import { type Order, orderService } from "@/services/order.service";
import { designSystemStyles } from "@/styles/design-system";

const orderDetailStyles = `
  @keyframes shimmer {
    0% { background-position: -200% 0; }
    100% { background-position: 200% 0; }
  }

  @keyframes pulse-soft {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.7; }
  }

  @keyframes float-subtle {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-4px); }
  }

  .progress-shimmer {
    background: linear-gradient(
      90deg,
      transparent 0%,
      rgba(255,255,255,0.4) 50%,
      transparent 100%
    );
    background-size: 200% 100%;
    animation: shimmer 2s infinite;
  }

  .status-step {
    transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .status-step:hover {
    transform: translateY(-2px);
  }

  .status-active {
    animation: pulse-soft 2s ease-in-out infinite;
  }

  .card-elevated {
    transition: all 0.3s ease;
  }

  .card-elevated:hover {
    transform: translateY(-2px);
    box-shadow: 0 20px 40px -12px rgba(61, 107, 90, 0.15);
  }

  .item-row {
    transition: background-color 0.2s ease;
  }

  .item-row:hover {
    background-color: rgba(126, 176, 155, 0.05);
  }

  .badge-float {
    animation: float-subtle 3s ease-in-out infinite;
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
    const baseClasses = "font-body font-medium px-5 py-2.5 text-sm tracking-wide badge-float";

    switch (status) {
      case "CREATED":
        return (
          <Badge className={`${baseClasses} bg-gradient-to-r from-amber-50 to-yellow-50 text-amber-700 border border-amber-200/60 shadow-sm`}>
            <Clock className="h-4 w-4 mr-2" />
            Aguardando
          </Badge>
        );
      case "PROCESSING":
        return (
          <Badge className={`${baseClasses} bg-gradient-to-r from-sky-50 to-blue-50 text-sky-700 border border-sky-200/60 shadow-sm`}>
            <Package className="h-4 w-4 mr-2" />
            Processando
          </Badge>
        );
      case "COMPLETED":
        return (
          <Badge className={`${baseClasses} bg-gradient-to-r from-emerald-50 to-teal-50 text-emerald-700 border border-emerald-200/60 shadow-sm`}>
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
        <style>{designSystemStyles}</style>
        <style>{orderDetailStyles}</style>
        <div className="min-h-screen bg-gradient-to-br from-[#F7FAF8] via-[#F5F9F7] to-[#F0F7F4]">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-16">
            <div className="flex flex-col justify-center items-center py-32">
              <div className="relative">
                <div className="absolute inset-0 bg-[#7EB09B]/20 blur-2xl rounded-full" />
                <Loader2 className="relative h-14 w-14 animate-spin text-[#5A9178]" />
              </div>
              <p className="font-body text-lg text-[#5A9178] mt-6 tracking-wide">
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
        <style>{designSystemStyles}</style>
        <style>{orderDetailStyles}</style>
        <div className="min-h-screen bg-gradient-to-br from-[#F7FAF8] via-[#F5F9F7] to-[#F0F7F4]">
          <Header />
          <main className="container mx-auto px-4 lg:px-8 py-16">
            <Card className="text-center py-20 rounded-3xl border border-[#7EB09B]/20 shadow-xl bg-white/80 backdrop-blur-sm">
              <CardContent>
                <div className="relative inline-block mb-8">
                  <div className="absolute inset-0 bg-[#7EB09B]/15 blur-3xl" />
                  <div className="relative bg-gradient-to-br from-[#7EB09B]/10 to-[#5A9178]/10 rounded-full p-8">
                    <Package className="h-16 w-16 text-[#5A9178]" />
                  </div>
                </div>
                <h3 className="font-display text-2xl font-semibold text-[#3D6B5A] mb-3">
                  Pedido não encontrado
                </h3>
                <p className="font-body text-[#5A9178] mb-8 max-w-md mx-auto">
                  O pedido que você está procurando não existe ou foi removido.
                </p>
                <Button
                  asChild
                  className="bg-gradient-to-r from-[#7EB09B] to-[#5A9178] hover:from-[#6A9D88] hover:to-[#4A8068] text-white font-body px-8 py-5 rounded-full text-sm shadow-lg shadow-[#7EB09B]/25 transition-all duration-300"
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
      <style>{designSystemStyles}</style>
      <style>{orderDetailStyles}</style>
      <div className="min-h-screen bg-gradient-to-br from-[#F7FAF8] via-[#F5F9F7] to-[#F0F7F4]">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-12">
          <div className={`mb-10 ${isVisible ? "hero-enter active" : "hero-enter"}`}>
            <Link
              to="/orders"
              className="inline-flex items-center font-body text-sm text-[#5A9178] hover:text-[#3D6B5A] transition-colors mb-6 group"
            >
              <ArrowLeft className="h-4 w-4 mr-2 group-hover:-translate-x-1 transition-transform" />
              Voltar para meus pedidos
            </Link>
            <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
              <div>
                <h1 className="font-display text-3xl lg:text-4xl font-semibold text-[#3D6B5A] mb-2 tracking-tight">
                  Pedido #{getOrderId(order).slice(0, 8)}
                </h1>
                <p className="font-body text-[#6B9080]">
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
              {/* Status Card */}
              <Card className="card-elevated shadow-lg border border-[#7EB09B]/15 rounded-2xl bg-white/90 backdrop-blur-sm overflow-hidden">
                <CardHeader className="pb-4">
                  <CardTitle className="font-display text-xl font-semibold text-[#3D6B5A] flex items-center gap-2">
                    <Sparkles className="h-5 w-5 text-[#7EB09B]" />
                    Acompanhamento
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-8">
                    {/* Progress Bar */}
                    <div className="relative">
                      <div className="overflow-hidden h-2 rounded-full bg-[#E8F0EC]">
                        <div
                          style={{ width: `${getStatusProgress(order.status)}%` }}
                          className="h-full rounded-full bg-gradient-to-r from-[#7EB09B] via-[#6A9D88] to-[#5A9178] transition-all duration-700 ease-out relative"
                        >
                          <div className="absolute inset-0 progress-shimmer rounded-full" />
                        </div>
                      </div>
                    </div>

                    {/* Status Steps */}
                    <div className="grid grid-cols-3 gap-3">
                      {/* Step 1: Created */}
                      <div
                        className={`status-step p-5 rounded-xl border ${
                          isStepActive(order.status, "CREATED")
                            ? "bg-gradient-to-br from-[#F0F7F4] to-[#E8F0EC] border-[#7EB09B]/40"
                            : "bg-[#FAFBFA] border-[#E5E7E5]"
                        } ${isCurrentStep(order.status, "CREATED") ? "status-active ring-2 ring-[#7EB09B]/20" : ""}`}
                      >
                        <div className={`w-10 h-10 rounded-full flex items-center justify-center mx-auto mb-3 ${
                          isStepActive(order.status, "CREATED")
                            ? "bg-gradient-to-br from-[#7EB09B] to-[#5A9178] shadow-md shadow-[#7EB09B]/30"
                            : "bg-[#E5E7E5]"
                        }`}>
                          <Clock className={`h-5 w-5 ${
                            isStepActive(order.status, "CREATED") ? "text-white" : "text-[#9CA3AF]"
                          }`} />
                        </div>
                        <p className={`font-body text-sm font-medium text-center ${
                          isStepActive(order.status, "CREATED") ? "text-[#3D6B5A]" : "text-[#9CA3AF]"
                        }`}>
                          Criado
                        </p>
                      </div>

                      {/* Step 2: Processing */}
                      <div
                        className={`status-step p-5 rounded-xl border ${
                          isStepActive(order.status, "PROCESSING")
                            ? "bg-gradient-to-br from-[#F0F7F4] to-[#E8F0EC] border-[#7EB09B]/40"
                            : "bg-[#FAFBFA] border-[#E5E7E5]"
                        } ${isCurrentStep(order.status, "PROCESSING") ? "status-active ring-2 ring-[#7EB09B]/20" : ""}`}
                      >
                        <div className={`w-10 h-10 rounded-full flex items-center justify-center mx-auto mb-3 ${
                          isStepActive(order.status, "PROCESSING")
                            ? "bg-gradient-to-br from-[#7EB09B] to-[#5A9178] shadow-md shadow-[#7EB09B]/30"
                            : "bg-[#E5E7E5]"
                        }`}>
                          <Package className={`h-5 w-5 ${
                            isStepActive(order.status, "PROCESSING") ? "text-white" : "text-[#9CA3AF]"
                          }`} />
                        </div>
                        <p className={`font-body text-sm font-medium text-center ${
                          isStepActive(order.status, "PROCESSING") ? "text-[#3D6B5A]" : "text-[#9CA3AF]"
                        }`}>
                          Processando
                        </p>
                      </div>

                      {/* Step 3: Completed */}
                      <div
                        className={`status-step p-5 rounded-xl border ${
                          isStepActive(order.status, "COMPLETED")
                            ? "bg-gradient-to-br from-[#F0F7F4] to-[#E8F0EC] border-[#7EB09B]/40"
                            : "bg-[#FAFBFA] border-[#E5E7E5]"
                        } ${isCurrentStep(order.status, "COMPLETED") ? "ring-2 ring-[#7EB09B]/20" : ""}`}
                      >
                        <div className={`w-10 h-10 rounded-full flex items-center justify-center mx-auto mb-3 ${
                          isStepActive(order.status, "COMPLETED")
                            ? "bg-gradient-to-br from-[#7EB09B] to-[#5A9178] shadow-md shadow-[#7EB09B]/30"
                            : "bg-[#E5E7E5]"
                        }`}>
                          <CheckCircle className={`h-5 w-5 ${
                            isStepActive(order.status, "COMPLETED") ? "text-white" : "text-[#9CA3AF]"
                          }`} />
                        </div>
                        <p className={`font-body text-sm font-medium text-center ${
                          isStepActive(order.status, "COMPLETED") ? "text-[#3D6B5A]" : "text-[#9CA3AF]"
                        }`}>
                          Concluído
                        </p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Items Card */}
              <Card className="card-elevated shadow-lg border border-[#7EB09B]/15 rounded-2xl bg-white/90 backdrop-blur-sm">
                <CardHeader className="pb-4">
                  <CardTitle className="font-display text-xl font-semibold text-[#3D6B5A]">
                    Itens do Pedido
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-1">
                    {order.items?.map((item, index) => (
                      <div key={index}>
                        <div className="item-row flex justify-between items-center py-4 px-3 -mx-3 rounded-xl">
                          <div>
                            <p className="font-display text-base font-medium text-[#3D6B5A]">
                              {item.name}
                            </p>
                            <p className="font-body text-sm text-[#6B9080] mt-0.5">
                              {item.quantity} {item.quantity > 1 ? "unidades" : "unidade"} × R$ {item.price.toFixed(2)}
                            </p>
                          </div>
                          <p className="font-display text-lg font-semibold text-[#5A9178]">
                            R$ {(item.price * item.quantity).toFixed(2)}
                          </p>
                        </div>
                        {index < (order.items?.length || 0) - 1 && (
                          <Separator className="bg-[#E8F0EC]" />
                        )}
                      </div>
                    )) || (
                      <p className="font-body text-[#6B9080] text-center py-8">
                        Nenhum item encontrado
                      </p>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Summary Card */}
            <div>
              <Card className="card-elevated shadow-lg border border-[#7EB09B]/15 rounded-2xl bg-white/90 backdrop-blur-sm sticky top-24 overflow-hidden">
                <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-[#7EB09B] via-[#6A9D88] to-[#5A9178]" />
                <CardHeader className="pb-3">
                  <CardTitle className="font-display text-xl font-semibold text-[#3D6B5A]">
                    Resumo
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex justify-between py-2 font-body">
                      <span className="text-[#6B9080]">Subtotal</span>
                      <span className="font-medium text-[#3D6B5A]">
                        R$ {order.total.toFixed(2)}
                      </span>
                    </div>
                    <div className="flex justify-between py-2 font-body">
                      <span className="text-[#6B9080]">Entrega</span>
                      <span className="font-medium text-[#7EB09B]">Grátis</span>
                    </div>
                    <Separator className="bg-[#E8F0EC]" />
                    <div className="flex justify-between items-baseline pt-2">
                      <span className="font-display font-semibold text-[#3D6B5A]">Total</span>
                      <span className="font-display text-2xl font-bold text-[#5A9178]">
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
