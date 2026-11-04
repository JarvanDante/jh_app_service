package public

import (
	"testing"

	"jh_app_service/internal/registry"
)

func TestFormatAmount(t *testing.T) {
	cases := []struct {
		v    float64
		d    int
		want string
	}{
		{1000, 0, "1000"}, {1000, 2, "1000.00"}, {12.345, 2, "12.35"}, {12.5, 0, "13"}, {5, -1, "5"},
	}
	for _, c := range cases {
		if got := formatAmount(c.v, c.d); got != c.want {
			t.Errorf("formatAmount(%v,%d)=%s want %s", c.v, c.d, got, c.want)
		}
	}
}

func TestSiteCurrencyFromConfig(t *testing.T) {
	if _, err := siteCurrencyFromConfig(&registry.SiteConfig{}); err == nil {
		t.Fatal("missing currency_config must fail")
	}
	cfg := &registry.SiteConfig{}
	cfg.Base.CurrencyConfig = registry.CurrencyConfig{Code: "MMK", Symbol: "K", SymbolPosition: 1, Decimals: 0, Locked: true}
	c, err := siteCurrencyFromConfig(cfg)
	if err != nil || c.Code != "MMK" || c.Decimals != 0 || !c.Locked {
		t.Fatalf("%v %v", c, err)
	}
}

func TestSiteLegacyCurrency(t *testing.T) {
	if _, err := siteLegacyCurrency(&registry.SiteConfig{}); err == nil {
		t.Fatal("empty currency must fail, no CNY fallback")
	}
	cfg := &registry.SiteConfig{}
	cfg.Base.Currency = "MMK(K)"
	if c, _ := siteLegacyCurrency(cfg); c != "MMK(K)" {
		t.Fatal(c)
	}
}
