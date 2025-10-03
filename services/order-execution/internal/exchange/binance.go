package exchange

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/b25/services/order-execution/internal/models"
)

const (
	BinanceFuturesBaseURL = "https://fapi.binance.com"
	BinanceFuturesTestURL = "https://testnet.binancefuture.com"
)

// BinanceClient handles Binance Futures API communication
type BinanceClient struct {
	apiKey     string
	secretKey  string
	baseURL    string
	httpClient *http.Client
	testnet    bool
}

// NewBinanceClient creates a new Binance Futures client
func NewBinanceClient(apiKey, secretKey string, testnet bool) *BinanceClient {
	baseURL := BinanceFuturesBaseURL
	if testnet {
		baseURL = BinanceFuturesTestURL
	}

	return &BinanceClient{
		apiKey:    apiKey,
		secretKey: secretKey,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		testnet: testnet,
	}
}

// CreateOrder creates a new order on Binance Futures
func (b *BinanceClient) CreateOrder(order *models.Order) (*BinanceOrderResponse, error) {
	params := make(map[string]string)
	params["symbol"] = order.Symbol
	params["side"] = string(order.Side)
	params["type"] = b.mapOrderType(order.Type)
	params["quantity"] = formatFloat(order.Quantity)

	// Set time in force
	if order.Type != models.OrderTypeMarket {
		params["timeInForce"] = b.mapTimeInForce(order.TimeInForce)
	}

	// Set price for limit orders
	if order.Type == models.OrderTypeLimit || order.Type == models.OrderTypeStopLimit || order.Type == models.OrderTypePostOnly {
		params["price"] = formatFloat(order.Price)
	}

	// Set stop price for stop orders
	if order.Type == models.OrderTypeStopMarket || order.Type == models.OrderTypeStopLimit {
		params["stopPrice"] = formatFloat(order.StopPrice)
	}

	// Set reduce only flag
	if order.ReduceOnly {
		params["reduceOnly"] = "true"
	}

	// Set post only flag
	if order.PostOnly || order.Type == models.OrderTypePostOnly {
		params["timeInForce"] = "GTX" // Good Till Crossing (Post-only)
	}

	// Set client order ID
	if order.ClientOrderID != "" {
		params["newClientOrderId"] = order.ClientOrderID
	}

	// Add timestamp
	params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Sign request
	signature := b.sign(params)
	params["signature"] = signature

	// Build query string
	queryString := b.buildQueryString(params)

	// Make request
	endpoint := "/fapi/v1/order"
	resp, err := b.post(endpoint, queryString)
	if err != nil {
		return nil, err
	}

	var orderResp BinanceOrderResponse
	if err := json.Unmarshal(resp, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order response: %w", err)
	}

	return &orderResp, nil
}

// CancelOrder cancels an order on Binance Futures
func (b *BinanceClient) CancelOrder(symbol, orderID string) (*BinanceCancelResponse, error) {
	params := make(map[string]string)
	params["symbol"] = symbol
	params["orderId"] = orderID
	params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)

	signature := b.sign(params)
	params["signature"] = signature

	queryString := b.buildQueryString(params)

	endpoint := "/fapi/v1/order"
	resp, err := b.delete(endpoint, queryString)
	if err != nil {
		return nil, err
	}

	var cancelResp BinanceCancelResponse
	if err := json.Unmarshal(resp, &cancelResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cancel response: %w", err)
	}

	return &cancelResp, nil
}

// GetOrder gets order information
func (b *BinanceClient) GetOrder(symbol, orderID string) (*BinanceOrderResponse, error) {
	params := make(map[string]string)
	params["symbol"] = symbol
	params["orderId"] = orderID
	params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)

	signature := b.sign(params)
	params["signature"] = signature

	queryString := b.buildQueryString(params)

	endpoint := "/fapi/v1/order"
	resp, err := b.get(endpoint, queryString)
	if err != nil {
		return nil, err
	}

	var orderResp BinanceOrderResponse
	if err := json.Unmarshal(resp, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order response: %w", err)
	}

	return &orderResp, nil
}

// GetExchangeInfo gets exchange trading rules and symbol information
func (b *BinanceClient) GetExchangeInfo() (*BinanceExchangeInfo, error) {
	endpoint := "/fapi/v1/exchangeInfo"
	resp, err := b.get(endpoint, "")
	if err != nil {
		return nil, err
	}

	var exchangeInfo BinanceExchangeInfo
	if err := json.Unmarshal(resp, &exchangeInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal exchange info: %w", err)
	}

	return &exchangeInfo, nil
}

// GetAccountInfo gets account information
func (b *BinanceClient) GetAccountInfo() (*BinanceAccountInfo, error) {
	params := make(map[string]string)
	params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)

	signature := b.sign(params)
	params["signature"] = signature

	queryString := b.buildQueryString(params)

	endpoint := "/fapi/v2/account"
	resp, err := b.get(endpoint, queryString)
	if err != nil {
		return nil, err
	}

	var accountInfo BinanceAccountInfo
	if err := json.Unmarshal(resp, &accountInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account info: %w", err)
	}

	return &accountInfo, nil
}

// sign creates HMAC SHA256 signature
func (b *BinanceClient) sign(params map[string]string) string {
	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query string
	var queryParts []string
	for _, k := range keys {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, url.QueryEscape(params[k])))
	}
	query := strings.Join(queryParts, "&")

	// Create signature
	h := hmac.New(sha256.New, []byte(b.secretKey))
	h.Write([]byte(query))
	return hex.EncodeToString(h.Sum(nil))
}

// buildQueryString builds URL query string from params
func (b *BinanceClient) buildQueryString(params map[string]string) string {
	var parts []string
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
	}
	return strings.Join(parts, "&")
}

// get makes GET request
func (b *BinanceClient) get(endpoint, queryString string) ([]byte, error) {
	url := b.baseURL + endpoint
	if queryString != "" {
		url += "?" + queryString
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", b.apiKey)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr BinanceAPIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("API error: %s", string(body))
		}
		return nil, &apiErr
	}

	return body, nil
}

// post makes POST request
func (b *BinanceClient) post(endpoint, queryString string) ([]byte, error) {
	url := b.baseURL + endpoint + "?" + queryString

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", b.apiKey)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr BinanceAPIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("API error: %s", string(body))
		}
		return nil, &apiErr
	}

	return body, nil
}

// delete makes DELETE request
func (b *BinanceClient) delete(endpoint, queryString string) ([]byte, error) {
	url := b.baseURL + endpoint + "?" + queryString

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", b.apiKey)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr BinanceAPIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("API error: %s", string(body))
		}
		return nil, &apiErr
	}

	return body, nil
}

// mapOrderType maps internal order type to Binance order type
func (b *BinanceClient) mapOrderType(orderType models.OrderType) string {
	switch orderType {
	case models.OrderTypeMarket:
		return "MARKET"
	case models.OrderTypeLimit, models.OrderTypePostOnly:
		return "LIMIT"
	case models.OrderTypeStopMarket:
		return "STOP_MARKET"
	case models.OrderTypeStopLimit:
		return "STOP"
	default:
		return "LIMIT"
	}
}

// mapTimeInForce maps internal TIF to Binance TIF
func (b *BinanceClient) mapTimeInForce(tif models.TimeInForce) string {
	switch tif {
	case models.TimeInForceGTC:
		return "GTC"
	case models.TimeInForceIOC:
		return "IOC"
	case models.TimeInForceFOK:
		return "FOK"
	case models.TimeInForceGTX:
		return "GTX"
	default:
		return "GTC"
	}
}

// formatFloat formats float to string with appropriate precision
func formatFloat(val float64) string {
	return strconv.FormatFloat(val, 'f', -1, 64)
}
