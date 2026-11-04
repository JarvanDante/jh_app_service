package public

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

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

// siteSettingCurrency returns the currency object for setting-site plus the
// decimals/rounding mode used for amount strings. When currency_config is
// missing the currency is nil and amounts fall back to 2 decimals HALF_UP;
// this is display-only and does not affect registration/balance/game.
func siteSettingCurrency(cfg *registry.SiteConfig) (*v1.SiteCurrency, int, string) {
	currency, err := siteCurrencyFromConfig(cfg)
	if err != nil {
		return nil, 2, RoundingHalfUp
	}
	return currency, int(currency.Decimals), currency.RoundingMode
}

const (
	RoundingHalfUp   = "HALF_UP"
	RoundingHalfEven = "HALF_EVEN"
	RoundingDown     = "DOWN"
	RoundingUp       = "UP"
)

// formatAmount renders an amount with HALF_UP rounding, see formatAmountMode.
func formatAmount(v float64, decimals int) string {
	return formatAmountMode(v, decimals, RoundingHalfUp)
}

// formatAmountMode renders an amount with the site's decimals as a plain number
// string (no thousand separators), e.g. "1000.00". Rounding is done on the
// shortest decimal representation of v (so 1.005 is treated as 1.005, not
// 1.00499999...) using big integers. Modes: HALF_UP / HALF_EVEN (ties away
// from zero / to even), DOWN (toward zero), UP (away from zero). Unknown or
// empty modes use HALF_UP.
func formatAmountMode(v float64, decimals int, mode string) string {
	if decimals < 0 {
		decimals = 0
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		v = 0
	}
	s := strconv.FormatFloat(v, 'f', -1, 64)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	intPart, frac := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, frac = s[:i], s[i+1:]
	}
	for len(frac) < decimals {
		frac += "0"
	}
	kept, rest := frac[:decimals], frac[decimals:]
	n, _ := new(big.Int).SetString(intPart+kept, 10)
	if n == nil {
		n = new(big.Int)
	}
	restNonZero := strings.Trim(rest, "0") != ""
	inc := false
	switch strings.ToUpper(strings.TrimSpace(mode)) {
	case RoundingDown:
	case RoundingUp:
		inc = restNonZero
	case RoundingHalfEven:
		if rest != "" {
			switch {
			case rest[0] > '5':
				inc = true
			case rest[0] == '5':
				if strings.Trim(rest[1:], "0") != "" {
					inc = true
				} else {
					inc = n.Bit(0) == 1
				}
			}
		}
	default:
		inc = rest != "" && rest[0] >= '5'
	}
	if inc {
		n.Add(n, big.NewInt(1))
	}
	digits := n.String()
	for len(digits) <= decimals {
		digits = "0" + digits
	}
	out := digits
	if decimals > 0 {
		out = digits[:len(digits)-decimals] + "." + digits[len(digits)-decimals:]
	}
	if neg && n.Sign() != 0 {
		out = "-" + out
	}
	return out
}

// siteLegacyCurrency returns the legacy currency value stored on users/balances
// (e.g. "MMK(K)"). It never falls back to a hard-coded currency.
func siteLegacyCurrency(cfg *registry.SiteConfig) (string, error) {
	if cfg == nil || cfg.Base.Currency == "" {
		return "", fmt.Errorf("站点未配置币种")
	}
	return cfg.Base.Currency, nil
}
