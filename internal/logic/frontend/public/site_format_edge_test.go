package public

import (
	"encoding/json"
	"testing"

	"jh_app_service/internal/registry"
)

// min/max_withdraw from the DB are formatted with the site decimals.
func TestFormatAmountWithdrawLimits(t *testing.T) {
	cases := []struct {
		v    float64
		d    int
		want string
	}{
		{1500, 2, "1500.00"}, {2000, 2, "2000.00"}, {1500, 0, "1500"}, {1500, 4, "1500.0000"},
		{0, 2, "0.00"}, {-2.5, 0, "-3"}, {1234567, 3, "1234567.000"},
	}
	for _, c := range cases {
		if got := formatAmount(c.v, c.d); got != c.want {
			t.Errorf("formatAmount(%v,%d)=%s want %s", c.v, c.d, got, c.want)
		}
	}
}

// A KV published before Phase 1 has no currency_config: setting-site must error rather than guess,
// while registration still works from the legacy Base.currency value.
func TestOldKVWithoutCurrencyConfig(t *testing.T) {
	raw := `{"Base":{"site_id":1,"code":"jh","timezone":"UTC","currency":"MMK(K)","status":1}}`
	var cfg registry.SiteConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := siteCurrencyFromConfig(&cfg); err == nil {
		t.Error("old KV without currency_config must fail")
	}
	if c, err := siteLegacyCurrency(&cfg); err != nil || c != "MMK(K)" {
		t.Errorf("legacy currency: %s %v", c, err)
	}
	if _, err := siteCurrencyFromConfig(nil); err == nil {
		t.Error("nil config must fail")
	}
	if _, err := siteLegacyCurrency(nil); err == nil {
		t.Error("nil config must fail")
	}
}

// The KV shape written by manage_service parses into the app registry types.
func TestNewKVParses(t *testing.T) {
	raw := `{"Base":{"site_id":1,"code":"jh","currency":"MMK(K)",
	"currency_config":{"code":"MMK","symbol":"K","symbol_position":1,"decimals":2,"thousand_sep":",","decimal_sep":".","rounding_mode":"HALF_UP","locked":true},
	"locale":{"locale":"en-US","default_lang":"en","supported_langs":["en","zh"],"phone_prefix":"+95"}}}`
	var cfg registry.SiteConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatal(err)
	}
	c, err := siteCurrencyFromConfig(&cfg)
	if err != nil || c.Code != "MMK" || c.Symbol != "K" || c.Decimals != 2 || c.ThousandSep != "," || c.RoundingMode != "HALF_UP" || !c.Locked {
		t.Fatalf("%+v %v", c, err)
	}
	if cfg.Base.Locale.PhonePrefix != "+95" || len(cfg.Base.Locale.SupportedLangs) != 2 || cfg.Base.Currency != "MMK(K)" {
		t.Errorf("locale/legacy: %+v", cfg.Base)
	}
}
