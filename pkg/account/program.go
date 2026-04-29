package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	ProgramSellingPolicyManagement = "SELLING_POLICY_MANAGEMENT"
	ProgramOutOfStockControl       = "OUT_OF_STOCK_CONTROL"
	ProgramPartnerMotorsDealer     = "PARTNER_MOTORS_DEALER"
	ProgramEbayPlus                = "EBAY_PLUS"
)

type Program struct {
	ProgramType string `json:"programType"`
}

type programList struct {
	Programs []Program `json:"programs"`
}

func (c *Client) GetOptedInPrograms(ctx context.Context) ([]Program, error) {
	res, err := c.doer.Do(ctx, http.MethodGet, "/program/get_opted_in_programs", nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("account: getOptedInPrograms %d: %s", res.StatusCode, string(res.Body))
	}
	var out programList
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("account: decode getOptedInPrograms: %w", err)
	}
	return out.Programs, nil
}

// OptIn enrolls the seller into a program. Already-opted-in is not an
// error; the call returns nil in that case so callers can treat it as
// idempotent. eBay surfaces "already enrolled" two ways: HTTP 409
// Conflict (newer), and HTTP 400 with errorId 20407 / 25804 in the
// body (older). Both are treated as success.
func (c *Client) OptIn(ctx context.Context, programType string) error {
	body, _ := json.Marshal(Program{ProgramType: programType})
	res, err := c.doer.Do(ctx, http.MethodPost, "/program/opt_in", bytes.NewReader(body), "application/json")
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNoContent {
		return nil
	}
	if res.StatusCode == http.StatusConflict || isAlreadyOptedIn(res.Body) {
		return nil
	}
	return fmt.Errorf("account: optIn %s %d: %s", programType, res.StatusCode, string(res.Body))
}

func isAlreadyOptedIn(body []byte) bool {
	var resp struct {
		Errors []struct {
			ErrorID int `json:"errorId"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return false
	}
	for _, e := range resp.Errors {
		if e.ErrorID == 20407 || e.ErrorID == 25804 {
			return true
		}
	}
	return false
}
