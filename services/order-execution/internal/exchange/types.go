package exchange

import "fmt"

// BinanceOrderResponse represents Binance order response
type BinanceOrderResponse struct {
	OrderID           int64   `json:"orderId"`
	Symbol            string  `json:"symbol"`
	Status            string  `json:"status"`
	ClientOrderID     string  `json:"clientOrderId"`
	Price             string  `json:"price"`
	AvgPrice          string  `json:"avgPrice"`
	OrigQty           string  `json:"origQty"`
	ExecutedQty       string  `json:"executedQty"`
	CumQty            string  `json:"cumQty"`
	CumQuote          string  `json:"cumQuote"`
	TimeInForce       string  `json:"timeInForce"`
	Type              string  `json:"type"`
	ReduceOnly        bool    `json:"reduceOnly"`
	ClosePosition     bool    `json:"closePosition"`
	Side              string  `json:"side"`
	PositionSide      string  `json:"positionSide"`
	StopPrice         string  `json:"stopPrice"`
	WorkingType       string  `json:"workingType"`
	PriceProtect      bool    `json:"priceProtect"`
	OrigType          string  `json:"origType"`
	UpdateTime        int64   `json:"updateTime"`
	ActivatePrice     string  `json:"activatePrice,omitempty"`
	PriceRate         string  `json:"priceRate,omitempty"`
}

// BinanceCancelResponse represents Binance cancel response
type BinanceCancelResponse struct {
	OrderID       int64  `json:"orderId"`
	Symbol        string `json:"symbol"`
	Status        string `json:"status"`
	ClientOrderID string `json:"clientOrderId"`
	Price         string `json:"price"`
	AvgPrice      string `json:"avgPrice"`
	OrigQty       string `json:"origQty"`
	ExecutedQty   string `json:"executedQty"`
	CumQty        string `json:"cumQty"`
	CumQuote      string `json:"cumQuote"`
	TimeInForce   string `json:"timeInForce"`
	Type          string `json:"type"`
	ReduceOnly    bool   `json:"reduceOnly"`
	Side          string `json:"side"`
	PositionSide  string `json:"positionSide"`
	StopPrice     string `json:"stopPrice"`
	UpdateTime    int64  `json:"updateTime"`
}

// BinanceExchangeInfo represents exchange information
type BinanceExchangeInfo struct {
	Timezone   string                  `json:"timezone"`
	ServerTime int64                   `json:"serverTime"`
	Symbols    []BinanceSymbolInfo     `json:"symbols"`
	RateLimits []BinanceRateLimitInfo  `json:"rateLimits"`
}

// BinanceSymbolInfo represents symbol trading rules
type BinanceSymbolInfo struct {
	Symbol                string                     `json:"symbol"`
	Status                string                     `json:"status"`
	BaseAsset             string                     `json:"baseAsset"`
	QuoteAsset            string                     `json:"quoteAsset"`
	PricePrecision        int                        `json:"pricePrecision"`
	QuantityPrecision     int                        `json:"quantityPrecision"`
	BaseAssetPrecision    int                        `json:"baseAssetPrecision"`
	QuotePrecision        int                        `json:"quotePrecision"`
	UnderlyingType        string                     `json:"underlyingType"`
	UnderlyingSubType     []string                   `json:"underlyingSubType,omitempty"`
	OrderTypes            []string                   `json:"orderTypes"`
	TimeInForce           []string                   `json:"timeInForce"`
	Filters               []BinanceSymbolFilter      `json:"filters"`
}

// BinanceSymbolFilter represents symbol filters
type BinanceSymbolFilter struct {
	FilterType          string `json:"filterType"`
	MinPrice            string `json:"minPrice,omitempty"`
	MaxPrice            string `json:"maxPrice,omitempty"`
	TickSize            string `json:"tickSize,omitempty"`
	MinQty              string `json:"minQty,omitempty"`
	MaxQty              string `json:"maxQty,omitempty"`
	StepSize            string `json:"stepSize,omitempty"`
	MinNotional         string `json:"minNotional,omitempty"`
	Limit               int    `json:"limit,omitempty"`
	MultiplierUp        string `json:"multiplierUp,omitempty"`
	MultiplierDown      string `json:"multiplierDown,omitempty"`
	MultiplierDecimal   string `json:"multiplierDecimal,omitempty"`
}

// BinanceRateLimitInfo represents rate limit information
type BinanceRateLimitInfo struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int    `json:"intervalNum"`
	Limit         int    `json:"limit"`
}

// BinanceAccountInfo represents account information
type BinanceAccountInfo struct {
	FeeTier                     int                         `json:"feeTier"`
	CanTrade                    bool                        `json:"canTrade"`
	CanDeposit                  bool                        `json:"canDeposit"`
	CanWithdraw                 bool                        `json:"canWithdraw"`
	UpdateTime                  int64                       `json:"updateTime"`
	TotalInitialMargin          string                      `json:"totalInitialMargin"`
	TotalMaintMargin            string                      `json:"totalMaintMargin"`
	TotalWalletBalance          string                      `json:"totalWalletBalance"`
	TotalUnrealizedProfit       string                      `json:"totalUnrealizedProfit"`
	TotalMarginBalance          string                      `json:"totalMarginBalance"`
	TotalPositionInitialMargin  string                      `json:"totalPositionInitialMargin"`
	TotalOpenOrderInitialMargin string                      `json:"totalOpenOrderInitialMargin"`
	TotalCrossWalletBalance     string                      `json:"totalCrossWalletBalance"`
	TotalCrossUnPnl             string                      `json:"totalCrossUnPnl"`
	AvailableBalance            string                      `json:"availableBalance"`
	MaxWithdrawAmount           string                      `json:"maxWithdrawAmount"`
	Assets                      []BinanceAccountAsset       `json:"assets"`
	Positions                   []BinanceAccountPosition    `json:"positions"`
}

// BinanceAccountAsset represents account asset balance
type BinanceAccountAsset struct {
	Asset                  string `json:"asset"`
	WalletBalance          string `json:"walletBalance"`
	UnrealizedProfit       string `json:"unrealizedProfit"`
	MarginBalance          string `json:"marginBalance"`
	MaintMargin            string `json:"maintMargin"`
	InitialMargin          string `json:"initialMargin"`
	PositionInitialMargin  string `json:"positionInitialMargin"`
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
	CrossWalletBalance     string `json:"crossWalletBalance"`
	CrossUnPnl             string `json:"crossUnPnl"`
	AvailableBalance       string `json:"availableBalance"`
	MaxWithdrawAmount      string `json:"maxWithdrawAmount"`
	MarginAvailable        bool   `json:"marginAvailable"`
	UpdateTime             int64  `json:"updateTime"`
}

// BinanceAccountPosition represents account position
type BinanceAccountPosition struct {
	Symbol                 string `json:"symbol"`
	InitialMargin          string `json:"initialMargin"`
	MaintMargin            string `json:"maintMargin"`
	UnrealizedProfit       string `json:"unrealizedProfit"`
	PositionInitialMargin  string `json:"positionInitialMargin"`
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
	Leverage               string `json:"leverage"`
	Isolated               bool   `json:"isolated"`
	EntryPrice             string `json:"entryPrice"`
	MaxNotional            string `json:"maxNotional"`
	PositionSide           string `json:"positionSide"`
	PositionAmt            string `json:"positionAmt"`
	UpdateTime             int64  `json:"updateTime"`
}

// BinanceAPIError represents Binance API error
type BinanceAPIError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *BinanceAPIError) Error() string {
	return fmt.Sprintf("Binance API error %d: %s", e.Code, e.Msg)
}

// MapBinanceOrderStatus maps Binance order status to internal status
func MapBinanceOrderStatus(status string) string {
	switch status {
	case "NEW":
		return "SUBMITTED"
	case "PARTIALLY_FILLED":
		return "PARTIALLY_FILLED"
	case "FILLED":
		return "FILLED"
	case "CANCELED":
		return "CANCELED"
	case "REJECTED":
		return "REJECTED"
	case "EXPIRED":
		return "EXPIRED"
	default:
		return "UNKNOWN"
	}
}
