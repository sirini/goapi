package models

import "testing"

func TestTradeEnumsRejectOutOfRangeValues(t *testing.T) {
	if TradePriceType(3).Valid() {
		t.Fatal("price type 3 must be rejected")
	}
	if TradeProductCondition(5).Valid() {
		t.Fatal("product condition 5 must be rejected")
	}
	if TradeShippingType(3).Valid() {
		t.Fatal("shipping type 3 must be rejected")
	}
	if TradeStatus(4).Valid() {
		t.Fatal("trade status 4 must be rejected")
	}
}

func TestTradeShippingLocationRequirement(t *testing.T) {
	if TRADE_SHIPPING_PARCEL.NeedsLocation() {
		t.Fatal("parcel delivery must not require a meetup location")
	}
	if !TRADE_SHIPPING_MEETUP.NeedsLocation() || !TRADE_SHIPPING_BOTH.NeedsLocation() {
		t.Fatal("meetup-capable delivery must require a location")
	}
}
