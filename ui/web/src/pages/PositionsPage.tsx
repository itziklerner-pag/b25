import { useTradingStore } from '@/store/trading';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { formatCurrency, formatNumber } from '@/lib/utils';
import { X } from 'lucide-react';
import { toast } from 'sonner';

export default function PositionsPage() {
  const positions = useTradingStore((state) => Array.from(state.positions.values()));
  const closePosition = useTradingStore((state) => state.closePosition);

  const handleClosePosition = (symbol: string) => {
    closePosition(symbol);
    toast.success(`Closing position for ${symbol}`);
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Open Positions ({positions.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {positions.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground">No open positions</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Symbol</TableHead>
                  <TableHead>Side</TableHead>
                  <TableHead className="text-right">Size</TableHead>
                  <TableHead className="text-right">Entry Price</TableHead>
                  <TableHead className="text-right">Mark Price</TableHead>
                  <TableHead className="text-right">Unrealized P&L</TableHead>
                  <TableHead className="text-right">Leverage</TableHead>
                  <TableHead className="text-right">Liq. Price</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {positions.map((position) => (
                  <TableRow key={position.symbol}>
                    <TableCell className="font-medium">{position.symbol}</TableCell>
                    <TableCell>
                      <span
                        className={`rounded px-2 py-1 text-xs font-medium ${
                          position.side === 'LONG'
                            ? 'bg-green-500/10 text-green-500'
                            : 'bg-red-500/10 text-red-500'
                        }`}
                      >
                        {position.side}
                      </span>
                    </TableCell>
                    <TableCell className="text-right">{formatNumber(position.size, 4)}</TableCell>
                    <TableCell className="text-right">{formatCurrency(position.entryPrice)}</TableCell>
                    <TableCell className="text-right">{formatCurrency(position.markPrice)}</TableCell>
                    <TableCell
                      className={`text-right font-medium ${
                        position.unrealizedPnl >= 0 ? 'text-green-500' : 'text-red-500'
                      }`}
                    >
                      {formatCurrency(position.unrealizedPnl)}
                    </TableCell>
                    <TableCell className="text-right">{position.leverage}x</TableCell>
                    <TableCell className="text-right">
                      {formatCurrency(position.liquidationPrice)}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleClosePosition(position.symbol)}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
