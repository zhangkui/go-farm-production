package domain

import (
	"errors"
	"testing"
)

// TestRecalculateBatchRemaining guards the input-batch quantity adjustment
// invariant: a batch whose quantity is lowered below its irreversible
// consumption (allocate - return + waste) must be rejected with a conflict,
// and waste must be counted as irreversible consumption (not ignored).
func TestRecalculateBatchRemaining(t *testing.T) {
	cases := []struct {
		name      string
		quantity  Decimal
		allocated Decimal
		returned  Decimal
		wasted    Decimal
		wantRem   Decimal
		wantErr   bool
	}{
		{
			// No ledger activity: remaining equals the received quantity.
			name:      "empty ledger",
			quantity:  100,
			allocated: 0,
			returned:  0,
			wasted:    0,
			wantRem:   100,
		},
		{
			// Allocate 10, return 2, waste 3 on a batch of 100:
			// irreversible = 10 - 2 + 3 = 11, remaining = 100 - 11 = 89.
			name:      "allocate return waste",
			quantity:  100,
			allocated: 10,
			returned:  2,
			wasted:    3,
			wantRem:   89,
		},
		{
			// Waste is irreversible and must not be ignored. If it were
			// dropped from the formula, remaining would wrongly read 92.
			name:      "waste counted as irreversible",
			quantity:  100,
			allocated: 10,
			returned:  2,
			wasted:    8,
			wantRem:   84,
		},
		{
			// Quantity exactly equals irreversible consumption: remaining 0,
			// no conflict (batch fully consumed, but legally so).
			name:      "quantity equals irreversible",
			quantity:  11,
			allocated: 10,
			returned:  2,
			wasted:    3,
			wantRem:   0,
		},
		{
			// Quantity lowered below irreversible consumption: over-consumed.
			// This is the reported bug — must return a conflict, not 0.
			name:      "quantity below irreversible",
			quantity:  5,
			allocated: 10,
			returned:  2,
			wasted:    3,
			wantErr:   true,
		},
		{
			// Quantity lowered below waste alone, even though net allocate
			// (10-2=8) is still under the quantity (10): waste makes the
			// irreversible consumption 11 > 10, so it must still conflict.
			name:      "quantity below waste",
			quantity:  10,
			allocated: 10,
			returned:  2,
			wasted:    3,
			wantErr:   true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rem, err := RecalculateBatchRemaining(c.quantity, c.allocated, c.returned, c.wasted)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected conflict error, got remaining=%v err=nil", rem)
				}
				var ae *AppError
				if !errors.As(err, &ae) || ae.HTTP != 409 || ae.Code != CodeConflict {
					t.Fatalf("expected 409 CodeConflict, got %v", err)
				}
				if rem != 0 {
					t.Fatalf("conflict should return zero remaining, got %v", rem)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rem != c.wantRem {
				t.Fatalf("remaining=%v want %v", rem, c.wantRem)
			}
		})
	}
}
