package arara

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// WalletService handles the /v1/wallet resource.
type WalletService struct {
	client *Client
}

// WalletTransactionDTO is a single wallet transaction.
type WalletTransactionDTO struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Description *string `json:"description"`
	ReferenceID *string `json:"referenceId"`
	Mode        string  `json:"mode"`
	CreatedAt   *string `json:"createdAt"`
}

// WalletTransactionPageDTO is a paginated page of wallet transactions.
type WalletTransactionPageDTO struct {
	Content       []WalletTransactionDTO `json:"content"`
	Page          int                    `json:"page"`
	Size          int                    `json:"size"`
	TotalElements int64                  `json:"totalElements"`
	TotalPages    int                    `json:"totalPages"`
}

// AutoRechargeSettingsDTO holds auto-recharge configuration.
type AutoRechargeSettingsDTO struct {
	Enabled           bool    `json:"enabled"`
	Threshold         float64 `json:"threshold"`
	Amount            float64 `json:"amount"`
	LastAttemptAt     *string `json:"lastAttemptAt"`
	LastFailureReason *string `json:"lastFailureReason"`
}

// UpdateAutoRechargeRequest is the payload for updating auto-recharge settings.
type UpdateAutoRechargeRequest struct {
	Enabled   *bool    `json:"enabled,omitempty"`
	Threshold *float64 `json:"threshold,omitempty"`
	Amount    *float64 `json:"amount,omitempty"`
}

// Transactions lists wallet transactions. GET /v1/wallet/transactions
func (s *WalletService) Transactions(ctx context.Context, page, size int) (*WalletTransactionPageDTO, error) {
	q := url.Values{
		"page": {strconv.Itoa(page)},
		"size": {strconv.Itoa(defaultSize(size, 20))},
	}
	var out WalletTransactionPageDTO
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/wallet/transactions", query: q}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAutoRecharge returns auto-recharge settings. GET /v1/wallet/auto-recharge
func (s *WalletService) GetAutoRecharge(ctx context.Context) (*AutoRechargeSettingsDTO, error) {
	var out AutoRechargeSettingsDTO
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/wallet/auto-recharge"}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAutoRecharge updates auto-recharge settings. PATCH /v1/wallet/auto-recharge
func (s *WalletService) UpdateAutoRecharge(ctx context.Context, req *UpdateAutoRechargeRequest) (*AutoRechargeSettingsDTO, error) {
	var out AutoRechargeSettingsDTO
	err := s.client.do(ctx, request{method: http.MethodPatch, path: "/v1/wallet/auto-recharge", body: req}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
