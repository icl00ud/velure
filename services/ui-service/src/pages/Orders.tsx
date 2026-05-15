import { CheckCircle, Clock, Loader2, Package } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import Header from "@/components/Header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "@/hooks/use-toast";
import { type Order, orderService } from "@/services/order.service";
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

  const loadOrders = useCallback(async () => {
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
        title: "Failed to load orders",
        description: error instanceof Error ? error.message : "Please try again later",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  }, [page]);

  useEffect(() => {
    loadOrders();
  }, [loadOrders]);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "CREATED":
        return (
          <Badge className="bg-yellow-100 text-yellow-700 border-2 border-yellow-300 font-body font-semibold px-4 py-1">
            <Clock className="h-4 w-4 mr-2" />
            Created
          </Badge>
        );
      case "PROCESSING":
        return (
          <Badge className="bg-blue-100 text-blue-700 border-2 border-blue-300 font-body font-semibold px-4 py-1">
            <Package className="h-4 w-4 mr-2" />
            Processing
          </Badge>
        );
      case "COMPLETED":
        return (
          <Badge className="bg-green-100 text-green-700 border-2 border-green-300 font-body font-semibold px-4 py-1">
            <CheckCircle className="h-4 w-4 mr-2" />
            Completed
          </Badge>
        );
      default:
        return <Badge className="font-body font-semibold">{status}</Badge>;
    }
  };

  const getOrderId = (order: Order) => order.id || order._id || "UNKNOWN";
  const getOrderDate = (order: Order) => {
    const dateStr = order.created_at || order.createdAt;
    if (!dateStr) return new Date();
    return new Date(dateStr);
  };

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-white">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-12">
          <div className={`mb-12 ${isVisible ? "hero-enter active" : "hero-enter"}`}>
            <span className="font-body text-[#52B788] font-semibold text-sm tracking-widest uppercase mb-4 block">
              My Account
            </span>
            <h1 className="font-display text-5xl lg:text-6xl font-bold text-[#1B4332] mb-4">
              My Orders
            </h1>
            <div className="w-20 h-1 bg-gradient-to-r from-[#52B788] to-[#A7C957] mb-6" />
            <p className="font-body text-xl text-[#2D6A4F]">Track the status of your orders</p>
          </div>

          {isLoading ? (
            <div className="flex flex-col justify-center items-center py-20">
              <Loader2 className="h-16 w-16 animate-spin text-[#52B788] mb-4" />
              <p className="font-body text-lg text-[#2D6A4F]">Loading orders...</p>
            </div>
          ) : !orders || orders.length === 0 ? (
            <Card className="text-center py-20 rounded-3xl border border-slate-200 shadow-2xl">
              <CardContent>
                <div className="relative inline-block mb-8">
                  <div className="absolute inset-0 bg-[#52B788]/20 blur-3xl" />
                  <div className="relative bg-gradient-to-br from-[#52B788]/10 to-[#95D5B2]/10 rounded-full p-8">
                    <Package className="h-20 w-20 text-[#52B788]" />
                  </div>
                </div>
                <h3 className="font-display text-3xl font-bold text-[#1B4332] mb-4">
                  No orders found
                </h3>
                <p className="font-body text-lg text-[#2D6A4F] mb-8 max-w-md mx-auto">
                  You haven't placed any orders yet. How about a little shopping?
                </p>
                <Button
                  asChild
                  className="btn-primary-custom font-body px-10 py-6 rounded-full text-lg"
                >
                  <Link to="/products">Start shopping</Link>
                </Button>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-6">
              {orders.map((order, _index) => {
                const orderId = getOrderId(order);
                return (
                  <Card
                    key={orderId}
                    className="shadow-lg border border-slate-200 rounded-3xl card-hover-subtle"
                  >
                    <CardHeader className="pb-4">
                      <div className="flex justify-between items-start">
                        <div>
                          <CardTitle className="font-display text-2xl font-bold text-[#1B4332]">
                            Order #{orderId.slice(0, 8)}
                          </CardTitle>
                          <p className="font-body text-sm text-[#2D6A4F] mt-2">
                            {getOrderDate(order).toLocaleDateString("en-US", {
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
                        <div className="flex justify-between items-center py-4 px-6 bg-slate-50 rounded-2xl">
                          <span className="font-body text-[#2D6A4F]">
                            {order.items?.length || 0}{" "}
                            {(order.items?.length || 0) === 1 ? "item" : "items"}
                          </span>
                          <span className="font-display text-2xl font-bold text-[#52B788]">
                            ${(order.total || 0).toFixed(2)}
                          </span>
                        </div>
                        <Button
                          asChild
                          className="w-full btn-primary-custom font-body text-lg rounded-full h-14"
                        >
                          <Link to={`/orders/${orderId}`}>View order details</Link>
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
                className="font-body border border-[#1B4332] hover:bg-[#1B4332] hover:text-white rounded-full px-6"
              >
                ← Previous
              </Button>
              <span className="font-body text-[#2D6A4F] px-4">
                Page <span className="font-bold text-[#52B788]">{page}</span> of{" "}
                <span className="font-bold text-[#52B788]">{totalPages}</span>
              </span>
              <Button
                variant="outline"
                disabled={page === totalPages}
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                className="font-body border border-[#1B4332] hover:bg-[#1B4332] hover:text-white rounded-full px-6"
              >
                Next →
              </Button>
            </div>
          )}
        </main>
      </div>
    </>
  );
};

export default Orders;
