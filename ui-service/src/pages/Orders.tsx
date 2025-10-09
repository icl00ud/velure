import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Clock, Package, CheckCircle, Loader2 } from "lucide-react";
import Header from "@/components/Header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { orderService, type Order } from "@/services/order.service";
import { toast } from "@/hooks/use-toast";

const Orders = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const pageSize = 10;

  useEffect(() => {
    loadOrders();
  }, [page]);

  const loadOrders = async () => {
    setIsLoading(true);
    try {
      const result = await orderService.getUserOrders(page, pageSize);
      setOrders(result.orders);
      setTotalPages(result.totalPages);
    } catch (error) {
      toast({
        title: "Erro ao carregar pedidos",
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
            <Clock className="h-3 w-3 mr-1" />
            Criado
          </Badge>
        );
      case "PROCESSING":
        return (
          <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-300">
            <Package className="h-3 w-3 mr-1" />
            Processando
          </Badge>
        );
      case "COMPLETED":
        return (
          <Badge variant="outline" className="bg-green-50 text-green-700 border-green-300">
            <CheckCircle className="h-3 w-3 mr-1" />
            Concluído
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <Header />

      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-foreground">Meus Pedidos</h1>
          <p className="text-muted-foreground mt-2">Acompanhe o status dos seus pedidos</p>
        </div>

        {isLoading ? (
          <div className="flex justify-center items-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : orders.length === 0 ? (
          <Card className="text-center py-12">
            <CardContent>
              <Package className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-xl font-semibold text-foreground mb-2">
                Nenhum pedido encontrado
              </h3>
              <p className="text-muted-foreground mb-6">
                Você ainda não fez nenhum pedido. Que tal começar a comprar?
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
          <div className="space-y-4">
            {orders.map((order) => (
              <Card key={order.id} className="shadow-soft hover:shadow-md transition-shadow">
                <CardHeader>
                  <div className="flex justify-between items-start">
                    <div>
                      <CardTitle className="text-lg">Pedido #{order.id.slice(0, 8)}</CardTitle>
                      <p className="text-sm text-muted-foreground mt-1">
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
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex justify-between items-center">
                      <span className="text-muted-foreground">
                        {order.items?.length || 0} {(order.items?.length || 0) === 1 ? "item" : "itens"}
                      </span>
                      <span className="text-lg font-semibold text-primary">
                        R$ {(order.total || 0).toFixed(2)}
                      </span>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        asChild
                        variant="outline"
                        className="flex-1"
                      >
                        <Link to={`/orders/${order.id}`}>Ver detalhes</Link>
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {totalPages > 1 && (
          <div className="flex justify-center gap-2 mt-8">
            <Button
              variant="outline"
              disabled={page === 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              Anterior
            </Button>
            <span className="flex items-center px-4">
              Página {page} de {totalPages}
            </span>
            <Button
              variant="outline"
              disabled={page === totalPages}
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            >
              Próxima
            </Button>
          </div>
        )}
      </main>
    </div>
  );
};

export default Orders;
