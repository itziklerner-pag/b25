import { useTradingStore } from '@/store/trading';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { formatCurrency, formatNumber, formatTimestamp } from '@/lib/utils';
import { X } from 'lucide-react';
import { toast } from 'sonner';

export default function OrdersPage() {
  const orders = useTradingStore((state) => Array.from(state.orders.values()));
  const cancelOrder = useTradingStore((state) => state.cancelOrder);

  const activeOrders = orders.filter((o) => o.status === 'NEW' || o.status === 'PARTIALLY_FILLED');
  const completedOrders = orders.filter((o) => o.status === 'FILLED' || o.status === 'CANCELED');

  const handleCancelOrder = (orderId: string) => {
    cancelOrder(orderId);
    toast.success('Order canceled');
  };

  const OrderTable = ({ orders, showActions }: { orders: typeof activeOrders; showActions: boolean }) => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Symbol</TableHead>
          <TableHead>Side</TableHead>
          <TableHead>Type</TableHead>
          <TableHead className="text-right">Price</TableHead>
          <TableHead className="text-right">Quantity</TableHead>
          <TableHead className="text-right">Filled</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="text-right">Time</TableHead>
          {showActions && <TableHead className="text-right">Actions</TableHead>}
        </TableRow>
      </TableHeader>
      <TableBody>
        {orders.map((order) => (
          <TableRow key={order.id}>
            <TableCell className="font-medium">{order.symbol}</TableCell>
            <TableCell>
              <span
                className={`rounded px-2 py-1 text-xs font-medium ${
                  order.side === 'BUY'
                    ? 'bg-green-500/10 text-green-500'
                    : 'bg-red-500/10 text-red-500'
                }`}
              >
                {order.side}
              </span>
            </TableCell>
            <TableCell>{order.type}</TableCell>
            <TableCell className="text-right">
              {order.price ? formatCurrency(order.price) : 'Market'}
            </TableCell>
            <TableCell className="text-right">{formatNumber(order.quantity, 4)}</TableCell>
            <TableCell className="text-right">{formatNumber(order.filledQuantity, 4)}</TableCell>
            <TableCell>
              <span
                className={`rounded px-2 py-1 text-xs font-medium ${
                  order.status === 'FILLED'
                    ? 'bg-green-500/10 text-green-500'
                    : order.status === 'CANCELED'
                      ? 'bg-red-500/10 text-red-500'
                      : 'bg-yellow-500/10 text-yellow-500'
                }`}
              >
                {order.status}
              </span>
            </TableCell>
            <TableCell className="text-right">{formatTimestamp(order.timestamp)}</TableCell>
            {showActions && (
              <TableCell className="text-right">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleCancelOrder(order.id)}
                >
                  <X className="h-4 w-4" />
                </Button>
              </TableCell>
            )}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Orders</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="active">
            <TabsList>
              <TabsTrigger value="active">Active ({activeOrders.length})</TabsTrigger>
              <TabsTrigger value="history">History ({completedOrders.length})</TabsTrigger>
            </TabsList>

            <TabsContent value="active">
              {activeOrders.length === 0 ? (
                <p className="text-center text-sm text-muted-foreground">No active orders</p>
              ) : (
                <OrderTable orders={activeOrders} showActions={true} />
              )}
            </TabsContent>

            <TabsContent value="history">
              {completedOrders.length === 0 ? (
                <p className="text-center text-sm text-muted-foreground">No order history</p>
              ) : (
                <OrderTable orders={completedOrders} showActions={false} />
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
