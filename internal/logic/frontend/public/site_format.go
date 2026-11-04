package public

import (
	"fmt"
	"math"
	"strconv"

	v1 "jh_app_service/api/frontend/public/v1"
	"jh_app_service/internal/registry"
)

// siteCurrencyFromConfig converts the published site config to the API object.
// The site must have been published with currency_config; otherwise an error is returned
// instead of guessing a currency.
func siteCurrencyFromConfig(cfg *registry.SiteConfig) (*v1.SiteCurrency, error) {
	if cfg == nil || cfg.Base.CurrencyConfig.Code == "" {
		return nil, fmt.Errorf("站点未配置币种，请在后台站点设置中配置币种后重新发布")
	}
	c := cfg.Base.CurrencyConfig
	return &v1.SiteCurrency{
		Code:           c.Code,
		Symbol:         c.Symbol,
		SymbolPosition: int32(c.SymbolPosition),
		Decimals:       int32(c.Decimals),
		ThousandSep:    c.ThousandSep,
		DecimalSep:     c.DecimalSep,
		RoundingMode:   c.RoundingMode,
		Locked:         c.Locked,
	}, nil
}

// formatAmount renders an amount with the site's decimals as a plain number
// string (no thousand separators), e.g. 1000 -> "1000" with 0 decimals, "1000.00" with 2 (MMK/USD).
func formatAmount(v float64, decimals int) string {
	if decimals < 0 {
		decimals = 0
	}
	p := math.Pow10(decimals)
	// half-up rounding away from zero
	r := math.Round(v*p) / p
	return strconv.FormatFloat(r, 'f', decimals, 64)
}

// siteLegacyCurrency returns the legacy currency value stored on users/balances
// (e.g. "MMK(K)"). It never falls back to a hard-coded currency.
func siteLegacyCurrency(cfg *registry.SiteConfig) (string, error) {
	if cfg == nil || cfg.Base.Currency == "" {
		return "", fmt.Errorf("站点未配置币种")
	}
	return cfg.Base.Currency, nil
}
