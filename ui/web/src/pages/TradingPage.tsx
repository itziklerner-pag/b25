import { useState, useMemo } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useTradingStore } from '@/store/trading';
import { formatCurrency } from '@/lib/utils';
import { toast } from 'sonner';
import type { OrderRequest } from '@/types';

export default function TradingPage() {
  const [side, setSide] = useState<'BUY' | 'SELL'>('BUY');
  const [type, setType] = useState<'LIMIT' | 'MARKET'>('LIMIT');
  const [price, setPrice] = useState('');
  const [quantity, setQuantity] = useState('');

  const account = useTradingStore((state) => state.account);
  const selectedSymbol = useTradingStore((state) => state.selectedSymbol);
  const sendOrder = useTradingStore((state) => state.sendOrder);

  const orderValue = useMemo(() => {
    const p = parseFloat(price) || 0;
    const q = parseFloat(quantity) || 0;
    return p * q;
  }, [price, quantity]);

  const validation = useMemo(() => {
    const p = parseFloat(price) || 0;
    const q = parseFloat(quantity) || 0;

    return {
      validPrice: type === 'MARKET' || p > 0,
      validQuantity: q > 0,
      sufficientBalance: account ? account.availableBalance >= orderValue : false,
    };
  }, [price, quantity, type, account, orderValue]);

  const isValid =
    validation.validPrice && validation.validQuantity && validation.sufficientBalance;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!isValid) {
      toast.error('Invalid order parameters');
      return;
    }

    const order: OrderRequest = {
      symbol: selectedSymbol,
      side,
      type,
      price: type === 'LIMIT' ? parseFloat(price) : undefined,
      quantity: parseFloat(quantity),
    };

    sendOrder(order);
    toast.success('Order submitted');

    // Reset form
    setPrice('');
    setQuantity('');
  };

  const setPercentage = (percentage: number) => {
    if (!account) return;

    const available = account.availableBalance;
    const p = parseFloat(price) || 0;

    if (p > 0) {
      const q = (available * percentage) / 100 / p;
      setQuantity(q.toFixed(8));
    }
  };

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Place Order</CardTitle>
          <CardDescription>Submit manual orders to the exchange</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Side Selection */}
            <div className="flex gap-2">
              <Button
                type="button"
                variant={side === 'BUY' ? 'default' : 'outline'}
                className={`flex-1 ${side === 'BUY' ? 'bg-green-600 hover:bg-green-700' : ''}`}
                onClick={() => setSide('BUY')}
              >
                Buy
              </Button>
              <Button
                type="button"
                variant={side === 'SELL' ? 'default' : 'outline'}
                className={`flex-1 ${side === 'SELL' ? 'bg-red-600 hover:bg-red-700' : ''}`}
                onClick={() => setSide('SELL')}
              >
                Sell
              </Button>
            </div>

            {/* Order Type */}
            <div className="flex gap-2">
              <Button
                type="button"
                variant={type === 'LIMIT' ? 'default' : 'outline'}
                className="flex-1"
                onClick={() => setType('LIMIT')}
              >
                Limit
              </Button>
              <Button
                type="button"
                variant={type === 'MARKET' ? 'default' : 'outline'}
                className="flex-1"
                onClick={() => setType('MARKET')}
              >
                Market
              </Button>
            </div>

            {/* Price */}
            {type === 'LIMIT' && (
              <div>
                <Label>Price</Label>
                <Input
                  type="number"
                  step="0.01"
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                  placeholder="0.00"
                />
              </div>
            )}

            {/* Quantity */}
            <div>
              <Label>Quantity</Label>
              <Input
                type="number"
                step="0.00000001"
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
                placeholder="0.00"
              />
              <div className="mt-2 flex gap-2">
                <Button type="button" size="sm" variant="outline" onClick={() => setPercentage(25)}>
                  25%
                </Button>
                <Button type="button" size="sm" variant="outline" onClick={() => setPercentage(50)}>
                  50%
                </Button>
                <Button type="button" size="sm" variant="outline" onClick={() => setPercentage(75)}>
                  75%
                </Button>
                <Button type="button" size="sm" variant="outline" onClick={() => setPercentage(100)}>
                  100%
                </Button>
              </div>
            </div>

            {/* Order Summary */}
            <div className="rounded-lg bg-muted p-3">
              <div className="flex justify-between text-sm">
                <span>Order Value:</span>
                <span className="font-medium">{formatCurrency(orderValue)}</span>
              </div>
              <div className="mt-1 flex justify-between text-sm">
                <span>Available:</span>
                <span className="font-medium">
                  {formatCurrency(account?.availableBalance || 0)}
                </span>
              </div>
            </div>

            {/* Submit Button */}
            <Button
              type="submit"
              disabled={!isValid}
              className="w-full"
              variant={side === 'BUY' ? 'default' : 'destructive'}
            >
              {side === 'BUY' ? 'Buy' : 'Sell'} {selectedSymbol}
            </Button>

            {!validation.sufficientBalance && (
              <p className="text-sm text-red-500">Insufficient balance</p>
            )}
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Account Information</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Balance:</span>
              <span className="font-medium">{formatCurrency(account?.balance || 0)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Equity:</span>
              <span className="font-medium">{formatCurrency(account?.equity || 0)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Available Balance:</span>
              <span className="font-medium">
                {formatCurrency(account?.availableBalance || 0)}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Unrealized P&L:</span>
              <span
                className={`font-medium ${
                  (account?.unrealizedPnl || 0) >= 0 ? 'text-green-500' : 'text-red-500'
                }`}
              >
                {formatCurrency(account?.unrealizedPnl || 0)}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
