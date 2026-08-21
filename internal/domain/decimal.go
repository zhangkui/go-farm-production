package domain

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

// Decimal is a fixed-point numeric type backed by float64.
// It maps to MySQL DECIMAL columns. The go-sql-driver/mysql returns DECIMAL
// values as []byte; we scan them into float64 to keep the codebase simple
// while preserving cent-level precision for farm-scale amounts.
type Decimal float64

// Scan implements sql.Scanner so DECIMAL columns ([]byte / string / float64)
// can be read into a Decimal.
func (d *Decimal) Scan(src any) error {
	if src == nil {
		*d = 0
		return nil
	}
	switch v := src.(type) {
	case []byte:
		f, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return fmt.Errorf("domain: cannot parse decimal %q: %w", string(v), err)
		}
		*d = Decimal(f)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("domain: cannot parse decimal %q: %w", v, err)
		}
		*d = Decimal(f)
	case float64:
		*d = Decimal(v)
	case float32:
		*d = Decimal(v)
	case int64:
		*d = Decimal(v)
	case nil:
		*d = 0
	default:
		return fmt.Errorf("domain: unsupported decimal source type %T", src)
	}
	return nil
}

// Value implements driver.Valuer so Decimals are written back as floats.
func (d Decimal) Value() (driver.Value, error) {
	return float64(d), nil
}

// Float64 returns the underlying float64 value.
func (d Decimal) Float64() float64 { return float64(d) }

// Add returns the sum of two decimals.
func (d Decimal) Add(other Decimal) Decimal { return Decimal(float64(d) + float64(other)) }

// Sub returns the difference of two decimals.
func (d Decimal) Sub(other Decimal) Decimal { return Decimal(float64(d) - float64(other)) }

// Mul returns the product of two decimals.
func (d Decimal) Mul(other Decimal) Decimal { return Decimal(float64(d) * float64(other)) }

// Div returns the quotient of two decimals.
func (d Decimal) Div(other Decimal) Decimal {
	if other == 0 {
		return 0
	}
	return Decimal(float64(d) / float64(other))
}

// IsZero reports whether the decimal equals zero.
func (d Decimal) IsZero() bool { return float64(d) == 0 }
