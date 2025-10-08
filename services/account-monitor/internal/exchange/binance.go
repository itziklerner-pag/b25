package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/account-monitor/internal/config"
	"github.com/yourorg/b25/services/account-monitor/internal/reconciliation"
)

type BinanceClient struct {
	config     config.ExchangeConfig
	httpClient *http.Client
	baseURL    string
	logger     *zap.Logger
	timeOffset int64 // Offset between server time and local time in milliseconds
}

func NewBinanceClient(cfg config.ExchangeConfig, logger *zap.Logger) *BinanceClient {
	futuresURL := "https://fapi.binance.com"
	if cfg.Testnet {
		futuresURL = "https://testnet.binancefuture.com"
	}

	client := &BinanceClient{
		config:     cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    futuresURL, // Use futures API by default for trading
		logger:     logger,
		timeOffset: 0,
	}

	// Synchronize time with Binance server
	if err := client.syncServerTime(context.Background()); err != nil {
		logger.Warn("Failed to sync server time, using local time", zap.Error(err))
	}

	return client
}

// syncServerTime synchronizes time with Binance server
func (b *BinanceClient) syncServerTime(ctx context.Context) error {
	endpoint := "/fapi/v1/time"

	req, err := http.NewRequestWithContext(ctx, "GET", b.baseURL+endpoint, nil)
	if err != nil {
		return err
	}

	localTime := time.Now().UnixMilli()

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var timeResp struct {
		ServerTime int64 `json:"serverTime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&timeResp); err != nil {
		return err
	}

	// Calculate offset: server time - local time
	b.timeOffset = timeResp.ServerTime - localTime

	b.logger.Info("Synchronized with Binance server time",
		zap.Int64("serverTime", timeResp.ServerTime),
		zap.Int64("localTime", localTime),
		zap.Int64("offset", b.timeOffset),
	)

	return nil
}

// getServerTime returns the current time adjusted to Binance server time
// Subtracts 1500ms safety margin to ensure we're never ahead of server time
func (b *BinanceClient) getServerTime() int64 {
	return time.Now().UnixMilli() + b.timeOffset - 1500
}

// getFreshServerTime fetches current server time from Binance
func (b *BinanceClient) getFreshServerTime(ctx context.Context) (int64, error) {
	endpoint := "/fapi/v1/time"

	req, err := http.NewRequestWithContext(ctx, "GET", b.baseURL+endpoint, nil)
	if err != nil {
		return 0, err
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var timeResp struct {
		ServerTime int64 `json:"serverTime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&timeResp); err != nil {
		return 0, err
	}

	return timeResp.ServerTime, nil
}

// GetAccountInfo fetches futures account information from Binance
func (b *BinanceClient) GetAccountInfo(ctx context.Context) (*reconciliation.ExchangeAccount, error) {
	// Get fresh server time for this request
	serverTime, err := b.getFreshServerTime(ctx)
	if err != nil {
		b.logger.Warn("Failed to get server time, using local time", zap.Error(err))
		serverTime = time.Now().UnixMilli()
	}

	// Use Futures API v2
	endpoint := "/fapi/v2/account"

	params := url.Values{}
	params.Add("timestamp", fmt.Sprintf("%d", serverTime))
	params.Add("recvWindow", "10000") // Allow 10 second window for timing variance

	signature := b.sign(params.Encode())
	params.Add("signature", signature)

	reqURL := fmt.Sprintf("%s%s?%s", b.baseURL, endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", b.config.APIKey)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance futures API error: %s", string(body))
	}

	// Futures account response structure
	var accountResp struct {
		TotalWalletBalance string `json:"totalWalletBalance"`
		AvailableBalance   string `json:"availableBalance"`
		Assets             []struct {
			Asset              string `json:"asset"`
			WalletBalance      string `json:"walletBalance"`
			UnrealizedProfit   string `json:"unrealizedProfit"`
			MarginBalance      string `json:"marginBalance"`
			AvailableBalance   string `json:"availableBalance"`
			CrossWalletBalance string `json:"crossWalletBalance"`
		} `json:"assets"`
		Positions []struct {
			Symbol           string `json:"symbol"`
			PositionAmt      string `json:"positionAmt"`
			EntryPrice       string `json:"entryPrice"`
			UnrealizedProfit string `json:"unrealizedProfit"`
			Leverage         string `json:"leverage"`
		} `json:"positions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&accountResp); err != nil {
		return nil, err
	}

	b.logger.Info("Fetched futures account info",
		zap.String("totalBalance", accountResp.TotalWalletBalance),
		zap.String("availableBalance", accountResp.AvailableBalance),
		zap.Int("assetCount", len(accountResp.Assets)),
		zap.Int("positionCount", len(accountResp.Positions)),
	)

	// Parse balances from assets
	balances := make(map[string]reconciliation.ExchangeBalance)
	for _, asset := range accountResp.Assets {
		walletBal, _ := decimal.NewFromString(asset.WalletBalance)
		availBal, _ := decimal.NewFromString(asset.AvailableBalance)
		unrealizedPnl, _ := decimal.NewFromString(asset.UnrealizedProfit)

		if walletBal.IsZero() && availBal.IsZero() {
			continue
		}

		// In futures, locked = wallet - available
		locked := walletBal.Sub(availBal)

		balances[asset.Asset] = reconciliation.ExchangeBalance{
			Free:   availBal,
			Locked: locked,
			Total:  walletBal.Add(unrealizedPnl), // Include unrealized P&L
		}
	}

	// Parse positions
	positions := make(map[string]reconciliation.ExchangePosition)
	for _, pos := range accountResp.Positions {
		qty, _ := decimal.NewFromString(pos.PositionAmt)
		if qty.IsZero() {
			continue
		}

		entryPrice, _ := decimal.NewFromString(pos.EntryPrice)
		positions[pos.Symbol] = reconciliation.ExchangePosition{
			Quantity:   qty,
			EntryPrice: entryPrice,
		}
	}

	return &reconciliation.ExchangeAccount{
		Balances:  balances,
		Positions: positions,
	}, nil
}

// sign creates HMAC SHA256 signature
func (b *BinanceClient) sign(payload string) string {
	mac := hmac.New(sha256.New, []byte(b.config.SecretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// GetListenKey gets user data stream listen key for futures
func (b *BinanceClient) GetListenKey(ctx context.Context) (string, error) {
	// Use Futures user data stream endpoint
	endpoint := "/fapi/v1/listenKey"

	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-MBX-APIKEY", b.config.APIKey)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get listen key: %s", string(body))
	}

	var result struct {
		ListenKey string `json:"listenKey"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	b.logger.Info("Got futures listen key", zap.String("listenKey", result.ListenKey[:10]+"..."))

	return result.ListenKey, nil
}

// KeepAliveListenKey keeps the futures listen key alive
func (b *BinanceClient) KeepAliveListenKey(ctx context.Context, listenKey string) error {
	// Use Futures user data stream endpoint
	endpoint := "/fapi/v1/listenKey"

	req, err := http.NewRequestWithContext(ctx, "PUT", b.baseURL+endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-MBX-APIKEY", b.config.APIKey)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to keep alive listen key: %s", string(body))
	}

	return nil
}
