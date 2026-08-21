package domain

import "testing"

// TestDecimalScan verifies DECIMAL column scanning from driver source types.
func TestDecimalScan(t *testing.T) {
	var d Decimal
	if err := d.Scan([]byte("12.34")); err != nil || d != 12.34 {
		t.Fatalf("scan []byte: d=%v err=%v", d, err)
	}
	if err := d.Scan("3.14"); err != nil || d != 3.14 {
		t.Fatalf("scan string: d=%v err=%v", d, err)
	}
	if err := d.Scan(nil); err != nil || d != 0 {
		t.Fatalf("scan nil: d=%v err=%v", d, err)
	}
	if err := d.Scan(42.0); err != nil || d != 42 {
		t.Fatalf("scan float64: d=%v err=%v", d, err)
	}
	if err := d.Scan(int64(7)); err != nil || d != 7 {
		t.Fatalf("scan int64: d=%v err=%v", d, err)
	}
	if err := d.Scan(struct{}{}); err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

// TestPaginationNormalize checks bounds coercion.
func TestPaginationNormalize(t *testing.T) {
	p := Pagination{Page: -1, PageSize: 999, Order: "sideways"}
	p.Normalize()
	if p.Page != DefaultPage || p.PageSize != MaxPageSize || p.Order != "desc" {
		t.Fatalf("normalize produced %+v", p)
	}
	if p.Offset() != 0 {
		t.Fatalf("offset=%d want 0", p.Offset())
	}
	p2 := Pagination{Page: 3, PageSize: 20}
	p2.Normalize()
	if p2.Offset() != 40 {
		t.Fatalf("offset=%d want 40", p2.Offset())
	}
}

// TestNewPageResult checks total-page computation and nil-slice init.
func TestNewPageResult(t *testing.T) {
	pr := NewPageResult[int](nil, 45, Pagination{Page: 1, PageSize: 20})
	if len(pr.Items) != 0 || pr.Total != 45 || pr.TotalPage != 3 {
		t.Fatalf("page result wrong: %+v", pr)
	}
}

// TestPlanStatusName covers the human-readable mapper.
func TestPlanStatusName(t *testing.T) {
	if PlanStatusName(PlanStatusPlanted) != "planted" {
		t.Fail()
	}
	if PlanStatusName(99) != "unknown" {
		t.Fail()
	}
}
