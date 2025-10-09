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

const OrderDetail = () => {
  const { id } = useParams<{ id: string }>();
  const [order, setOrder] = useState<Order | null>(null);
  const [isLoading, setIsLoading] = useState(true);

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
          <Badge variant="outline" className="bg-yellow-50 text-yellow-700 border-yellow-300">
            <Clock className="h-4 w-4 mr-2" />
            Criado
          </Badge>
        );
      case "PROCESSING":
        return (
          <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-300">
            <Package className="h-4 w-4 mr-2" />
            Processando
          </Badge>
        );
      case "COMPLETED":
        return (
          <Badge variant="outline" className="bg-green-50 text-green-700 border-green-300">
            <CheckCircle className="h-4 w-4 mr-2" />
            Concluído
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
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
      <div className="min-h-screen bg-background">
        <Header />
        <main className="container mx-auto px-4 py-8">
          <div className="flex justify-center items-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        </main>
      </div>
    );
  }

  if (!order) {
    return (
      <div className="min-h-screen bg-background">
        <Header />
        <main className="container mx-auto px-4 py-8">
          <Card className="text-center py-12">
            <CardContent>
              <Package className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-xl font-semibold text-foreground mb-2">Pedido não encontrado</h3>
              <p className="text-muted-foreground mb-6">
                O pedido que você está procurando não existe ou foi removido.
              </p>
              <Button asChild className="bg-gradient-primary hover:opacity-90 text-primary-foreground">
                <Link to="/orders">Voltar para meus pedidos</Link>
              </Button>
            </CardContent>
          </Card>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <Header />

      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <Link
            to="/orders"
            className="inline-flex items-center text-muted-foreground hover:text-primary transition-colors mb-4"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Voltar para meus pedidos
          </Link>
          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-4xl font-bold text-foreground">Pedido #{order.id.slice(0, 8)}</h1>
              <p className="text-muted-foreground mt-2">
                Realizado em{" "}
                {new Date(order.created_at).toLocaleDateString("pt-BR", {
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
          <div className="lg:col-span-2 space-y-6">
            <Card className="shadow-soft">
              <CardHeader>
                <CardTitle>Status do Pedido</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="relative pt-1">
                    <div className="overflow-hidden h-2 text-xs flex rounded bg-gray-200">
                      <div
                        style={{ width: `${getStatusProgress(order.status)}%` }}
                        className="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-gradient-primary transition-all duration-500"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-3 gap-4 text-center">
                    <div
                      className={`p-3 rounded-lg ${
                        order.status === "CREATED" ||
                        order.status === "PROCESSING" ||
                        order.status === "COMPLETED"
                          ? "bg-primary/10"
                          : "bg-gray-100"
                      }`}
                    >
                      <Clock
                        className={`h-6 w-6 mx-auto mb-2 ${
                          order.status === "CREATED" ||
                          order.status === "PROCESSING" ||
                          order.status === "COMPLETED"
                            ? "text-primary"
                            : "text-gray-400"
                        }`}
                      />
                      <p className="text-xs font-medium">Criado</p>
                    </div>
                    <div
                      className={`p-3 rounded-lg ${
                        order.status === "PROCESSING" || order.status === "COMPLETED"
                          ? "bg-primary/10"
                          : "bg-gray-100"
                      }`}
                    >
                      <Package
                        className={`h-6 w-6 mx-auto mb-2 ${
                          order.status === "PROCESSING" || order.status === "COMPLETED"
                            ? "text-primary"
                            : "text-gray-400"
                        }`}
                      />
                      <p className="text-xs font-medium">Processando</p>
                    </div>
                    <div
                      className={`p-3 rounded-lg ${
                        order.status === "COMPLETED" ? "bg-primary/10" : "bg-gray-100"
                      }`}
                    >
                      <CheckCircle
                        className={`h-6 w-6 mx-auto mb-2 ${
                          order.status === "COMPLETED" ? "text-primary" : "text-gray-400"
                        }`}
                      />
                      <p className="text-xs font-medium">Concluído</p>
                    </div>
                  </div>

                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <p className="text-sm text-blue-800">
                      ℹ️ O status do seu pedido é atualizado automaticamente em tempo real.
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="shadow-soft">
              <CardHeader>
                <CardTitle>Itens do Pedido</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {order.items.map((item, index) => (
                    <div key={index}>
                      <div className="flex justify-between items-start">
                        <div>
                          <p className="font-medium">{item.name}</p>
                          <p className="text-sm text-muted-foreground">
                            Quantidade: {item.quantity}
                          </p>
                        </div>
                        <div className="text-right">
                          <p className="font-semibold">
                            R$ {(item.price * item.quantity).toFixed(2)}
                          </p>
                          <p className="text-sm text-muted-foreground">
                            R$ {item.price.toFixed(2)} cada
                          </p>
                        </div>
                      </div>
                      {index < order.items.length - 1 && <Separator className="mt-4" />}
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>

          <div>
            <Card className="shadow-soft">
              <CardHeader>
                <CardTitle>Resumo do Pedido</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Subtotal</span>
                    <span className="font-medium">R$ {order.total.toFixed(2)}</span>
                  </div>
                  <Separator />
                  <div className="flex justify-between text-lg font-semibold">
                    <span>Total</span>
                    <span className="text-primary">R$ {order.total.toFixed(2)}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  );
};

export default OrderDetail;
