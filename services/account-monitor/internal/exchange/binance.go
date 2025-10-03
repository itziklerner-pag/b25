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
}

func NewBinanceClient(cfg config.ExchangeConfig, logger *zap.Logger) *BinanceClient {
	baseURL := "https://api.binance.com"
	if cfg.Testnet {
		baseURL = "https://testnet.binance.vision"
	}

	return &BinanceClient{
		config:     cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		logger:     logger,
	}
}

// GetAccountInfo fetches account information from Binance
func (b *BinanceClient) GetAccountInfo(ctx context.Context) (*reconciliation.ExchangeAccount, error) {
	endpoint := "/api/v3/account"

	params := url.Values{}
	params.Add("timestamp", fmt.Sprintf("%d", time.Now().UnixMilli()))

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
		return nil, fmt.Errorf("binance API error: %s", string(body))
	}

	var accountResp struct {
		Balances []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
		Positions []struct {
			Symbol     string `json:"symbol"`
			PositionAmt string `json:"positionAmt"`
			EntryPrice string `json:"entryPrice"`
		} `json:"positions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&accountResp); err != nil {
		return nil, err
	}

	// Parse balances
	balances := make(map[string]reconciliation.ExchangeBalance)
	for _, bal := range accountResp.Balances {
		free, _ := decimal.NewFromString(bal.Free)
		locked, _ := decimal.NewFromString(bal.Locked)

		if free.IsZero() && locked.IsZero() {
			continue
		}

		balances[bal.Asset] = reconciliation.ExchangeBalance{
			Free:   free,
			Locked: locked,
			Total:  free.Add(locked),
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

// GetListenKey gets user data stream listen key
func (b *BinanceClient) GetListenKey(ctx context.Context) (string, error) {
	endpoint := "/api/v3/userDataStream"

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

	var result struct {
		ListenKey string `json:"listenKey"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ListenKey, nil
}

// KeepAliveListenKey keeps the listen key alive
func (b *BinanceClient) KeepAliveListenKey(ctx context.Context, listenKey string) error {
	endpoint := "/api/v3/userDataStream"

	params := url.Values{}
	params.Add("listenKey", listenKey)

	reqURL := fmt.Sprintf("%s%s?%s", b.baseURL, endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "PUT", reqURL, nil)
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
