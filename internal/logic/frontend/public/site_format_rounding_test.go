package public

import (
	"testing"

	"jh_app_service/internal/registry"
)

func TestFormatAmountModes(t *testing.T) {
	cases := []struct {
		v    float64
		d    int
		mode string
		want string
	}{
		{1.005, 2, RoundingHalfUp, "1.01"},
		{2.675, 2, RoundingHalfUp, "2.68"},
		{1.005, 2, "", "1.01"},
		{1.005, 2, "bogus", "1.01"},
		{1.004, 2, RoundingHalfUp, "1.00"},
		{1.005, 2, RoundingHalfEven, "1.00"},
		{1.015, 2, RoundingHalfEven, "1.02"},
		{1.0051, 2, RoundingHalfEven, "1.01"},
		{2.5, 0, RoundingHalfEven, "2"},
		{3.5, 0, RoundingHalfEven, "4"},
		{1.009, 2, RoundingDown, "1.00"},
		{1.001, 2, RoundingUp, "1.01"},
		{1.000, 2, RoundingUp, "1.00"},
		{-1.005, 2, RoundingHalfUp, "-1.01"},
		{-1.009, 2, RoundingDown, "-1.00"},
		{-0.001, 2, RoundingHalfUp, "0.00"},
		{0.999, 2, RoundingHalfUp, "1.00"},
		{9.995, 2, RoundingHalfUp, "10.00"},
		{1000, 0, RoundingHalfUp, "1000"},
		{1000, 2, RoundingHalfUp, "1000.00"},
		{50000000, 4, RoundingHalfUp, "50000000.0000"},
		{0.5, 0, "half_up", "1"},
	}
	for _, c := range cases {
		if got := formatAmountMode(c.v, c.d, c.mode); got != c.want {
			t.Errorf("formatAmountMode(%v,%d,%q)=%s want %s", c.v, c.d, c.mode, got, c.want)
		}
	}
}

func TestSiteSettingCurrencyMissingConfig(t *testing.T) {
	cur, d, mode := siteSettingCurrency(&registry.SiteConfig{})
	if cur != nil || d != 2 || mode != RoundingHalfUp {
		t.Fatalf("missing currency_config: cur=%v d=%d mode=%s", cur, d, mode)
	}
	cur, d, _ = siteSettingCurrency(nil)
	if cur != nil || d != 2 {
		t.Fatalf("nil cfg: cur=%v d=%d", cur, d)
	}
	cfg := &registry.SiteConfig{}
	cfg.Base.CurrencyConfig.Code = "MMK"
	cfg.Base.CurrencyConfig.Decimals = 0
	cfg.Base.CurrencyConfig.RoundingMode = RoundingDown
	cur, d, mode = siteSettingCurrency(cfg)
	if cur == nil || d != 0 || mode != RoundingDown {
		t.Fatalf("configured: cur=%v d=%d mode=%s", cur, d, mode)
	}
	// legacy registration path keeps erroring without a currency
	if _, err := siteLegacyCurrency(&registry.SiteConfig{}); err == nil {
		t.Fatal("registration must still have no fallback currency")
	}
}
