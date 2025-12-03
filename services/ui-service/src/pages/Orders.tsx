import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Clock, Package, CheckCircle, Loader2 } from "lucide-react";
import Header from "@/components/Header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { orderService, type Order } from "@/services/order.service";
import { toast } from "@/hooks/use-toast";
import { designSystemStyles } from "@/styles/design-system";

const Orders = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [isVisible, setIsVisible] = useState(false);
  const pageSize = 10;

  useEffect(() => {
    setIsVisible(true);
  }, []);

  useEffect(() => {
    loadOrders();
  }, [page]);

  const loadOrders = async () => {
    setIsLoading(true);
    try {
      const result = await orderService.getUserOrders(page, pageSize);
      
      // Handle different response formats
      let ordersList: Order[] = [];
      if (Array.isArray(result)) {
        ordersList = result;
      } else if (result?.orders) {
        ordersList = result.orders;
      } else if (result?.items) {
        ordersList = result.items;
      } else if (result?.data) {
        ordersList = result.data;
      }
      
      setOrders(ordersList || []);
      
      // Handle pagination
      if (result?.totalPages) {
        setTotalPages(result.totalPages);
      } else if (result?.totalCount) {
        setTotalPages(Math.ceil(result.totalCount / pageSize));
      } else {
        setTotalPages(1);
      }
    } catch (error) {
      setOrders([]);
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
          <Badge className="bg-yellow-100 text-yellow-700 border-2 border-yellow-300 font-body font-semibold px-4 py-1">
            <Clock className="h-4 w-4 mr-2" />
            Criado
          </Badge>
        );
      case "PROCESSING":
        return (
          <Badge className="bg-blue-100 text-blue-700 border-2 border-blue-300 font-body font-semibold px-4 py-1">
            <Package className="h-4 w-4 mr-2" />
            Processando
          </Badge>
        );
      case "COMPLETED":
        return (
          <Badge className="bg-green-100 text-green-700 border-2 border-green-300 font-body font-semibold px-4 py-1">
            <CheckCircle className="h-4 w-4 mr-2" />
            Concluído
          </Badge>
        );
      default:
        return <Badge className="font-body font-semibold">{status}</Badge>;
    }
  };

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
            <span className="font-body text-[#D97757] font-semibold text-sm tracking-widest uppercase mb-4 block">
              Minha Conta
            </span>
            <h1 className="font-display text-5xl lg:text-6xl font-bold text-[#2D3319] mb-4">
              Meus Pedidos
            </h1>
            <div className="w-20 h-1 bg-gradient-to-r from-[#D97757] to-[#F4C430] mb-6" />
            <p className="font-body text-xl text-[#5A6751]">
              Acompanhe o status dos seus pedidos
            </p>
          </div>

          {isLoading ? (
            <div className="flex flex-col justify-center items-center py-20">
              <Loader2 className="h-16 w-16 animate-spin text-[#D97757] mb-4" />
              <p className="font-body text-lg text-[#5A6751]">Carregando pedidos...</p>
            </div>
          ) : !orders || orders.length === 0 ? (
            <Card className="text-center py-20 rounded-3xl border-2 border-[#2D3319]/10 shadow-2xl">
              <CardContent>
                <div className="relative inline-block mb-8">
                  <div className="absolute inset-0 bg-[#D97757]/20 blur-3xl" />
                  <div className="relative bg-gradient-to-br from-[#D97757]/10 to-[#8B9A7E]/10 rounded-full p-8">
                    <Package className="h-20 w-20 text-[#D97757]" />
                  </div>
                </div>
                <h3 className="font-display text-3xl font-bold text-[#2D3319] mb-4">
                  Nenhum pedido encontrado
                </h3>
                <p className="font-body text-lg text-[#5A6751] mb-8 max-w-md mx-auto">
                  Você ainda não fez nenhum pedido. Que tal começar a comprar?
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
            <div className="space-y-6">
              {orders.map((order, index) => {
                const orderId = getOrderId(order);
                return (
                  <Card
                    key={orderId}
                    className="shadow-lg border-2 border-[#2D3319]/10 rounded-3xl card-hover-subtle"
                  >
                    <CardHeader className="pb-4">
                      <div className="flex justify-between items-start">
                        <div>
                          <CardTitle className="font-display text-2xl font-bold text-[#2D3319]">
                            Pedido #{orderId.slice(0, 8)}
                          </CardTitle>
                          <p className="font-body text-sm text-[#5A6751] mt-2">
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
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        <div className="flex justify-between items-center py-4 px-6 bg-[#FAF7F2] rounded-2xl">
                          <span className="font-body text-[#5A6751]">
                            {order.items?.length || 0}{" "}
                            {(order.items?.length || 0) === 1 ? "item" : "itens"}
                          </span>
                          <span className="font-display text-2xl font-bold text-[#D97757]">
                            R$ {(order.total || 0).toFixed(2)}
                          </span>
                        </div>
                        <Button
                          asChild
                          className="w-full btn-primary-custom font-body text-lg rounded-full h-14"
                        >
                          <Link to={`/orders/${orderId}`}>Ver detalhes do pedido</Link>
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
            </div>
          )}

          {totalPages > 1 && (
            <div className="flex justify-center items-center gap-4 mt-16">
              <Button
                variant="outline"
                disabled={page === 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="font-body border-2 border-[#2D3319] hover:bg-[#2D3319] hover:text-white rounded-full px-6"
              >
                ← Anterior
              </Button>
              <span className="font-body text-[#5A6751] px-4">
                Página <span className="font-bold text-[#D97757]">{page}</span> de{" "}
                <span className="font-bold text-[#D97757]">{totalPages}</span>
              </span>
              <Button
                variant="outline"
                disabled={page === totalPages}
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                className="font-body border-2 border-[#2D3319] hover:bg-[#2D3319] hover:text-white rounded-full px-6"
              >
                Próxima →
              </Button>
            </div>
          )}
        </main>
      </div>
    </>
  );
};

export default Orders;
