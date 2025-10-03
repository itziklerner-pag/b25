package types

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"strings"
)

// Decimal represents a high-precision decimal number for financial calculations.
// Uses big.Rat internally to avoid floating point errors.
type Decimal struct {
	value *big.Rat
}

// NewDecimal creates a new Decimal from a string representation.
func NewDecimal(value string) (*Decimal, error) {
	r := new(big.Rat)
	_, ok := r.SetString(value)
	if !ok {
		return nil, fmt.Errorf("invalid decimal string: %s", value)
	}
	return &Decimal{value: r}, nil
}

// NewDecimalFromInt64 creates a new Decimal from an int64.
func NewDecimalFromInt64(value int64) *Decimal {
	return &Decimal{value: big.NewRat(value, 1)}
}

// NewDecimalFromFloat creates a new Decimal from a float64.
// Note: This may introduce floating point errors. Use NewDecimal for exact values.
func NewDecimalFromFloat(value float64) *Decimal {
	r := new(big.Rat)
	r.SetFloat64(value)
	return &Decimal{value: r}
}

// Zero returns a decimal representing 0.
func Zero() *Decimal {
	return &Decimal{value: big.NewRat(0, 1)}
}

// String returns the string representation of the decimal.
func (d *Decimal) String() string {
	if d.value == nil {
		return "0"
	}
	return d.value.FloatString(18) // 18 decimal places precision
}

// Float64 returns the float64 representation (may lose precision).
func (d *Decimal) Float64() float64 {
	if d.value == nil {
		return 0
	}
	f, _ := d.value.Float64()
	return f
}

// Add returns d + other.
func (d *Decimal) Add(other *Decimal) *Decimal {
	result := new(big.Rat)
	result.Add(d.value, other.value)
	return &Decimal{value: result}
}

// Sub returns d - other.
func (d *Decimal) Sub(other *Decimal) *Decimal {
	result := new(big.Rat)
	result.Sub(d.value, other.value)
	return &Decimal{value: result}
}

// Mul returns d * other.
func (d *Decimal) Mul(other *Decimal) *Decimal {
	result := new(big.Rat)
	result.Mul(d.value, other.value)
	return &Decimal{value: result}
}

// Div returns d / other.
func (d *Decimal) Div(other *Decimal) *Decimal {
	result := new(big.Rat)
	result.Quo(d.value, other.value)
	return &Decimal{value: result}
}

// Cmp compares d and other. Returns -1 if d < other, 0 if equal, 1 if d > other.
func (d *Decimal) Cmp(other *Decimal) int {
	return d.value.Cmp(other.value)
}

// IsZero returns true if the decimal is zero.
func (d *Decimal) IsZero() bool {
	return d.value.Sign() == 0
}

// IsPositive returns true if the decimal is positive.
func (d *Decimal) IsPositive() bool {
	return d.value.Sign() > 0
}

// IsNegative returns true if the decimal is negative.
func (d *Decimal) IsNegative() bool {
	return d.value.Sign() < 0
}

// Abs returns the absolute value.
func (d *Decimal) Abs() *Decimal {
	result := new(big.Rat)
	result.Abs(d.value)
	return &Decimal{value: result}
}

// Neg returns the negation.
func (d *Decimal) Neg() *Decimal {
	result := new(big.Rat)
	result.Neg(d.value)
	return &Decimal{value: result}
}

// Round rounds to the specified number of decimal places.
func (d *Decimal) Round(places int) *Decimal {
	if places < 0 {
		places = 0
	}

	multiplier := big.NewRat(1, 1)
	for i := 0; i < places; i++ {
		multiplier.Mul(multiplier, big.NewRat(10, 1))
	}

	result := new(big.Rat)
	result.Mul(d.value, multiplier)

	// Round to nearest integer
	num := result.Num()
	denom := result.Denom()
	div := new(big.Int).Div(num, denom)
	rem := new(big.Int).Rem(num, denom)

	// Add 1 if remainder >= 0.5
	half := new(big.Int).Div(denom, big.NewInt(2))
	if rem.Cmp(half) >= 0 {
		div.Add(div, big.NewInt(1))
	}

	result.SetInt(div)
	result.Quo(result, multiplier)

	return &Decimal{value: result}
}

// Truncate truncates to the specified number of decimal places.
func (d *Decimal) Truncate(places int) *Decimal {
	if places < 0 {
		places = 0
	}

	str := d.String()
	parts := strings.Split(str, ".")

	if len(parts) == 1 || places == 0 {
		dec, _ := NewDecimal(parts[0])
		return dec
	}

	if len(parts[1]) <= places {
		return d
	}

	truncated := parts[0] + "." + parts[1][:places]
	dec, _ := NewDecimal(truncated)
	return dec
}

// Value implements driver.Valuer for database storage.
func (d *Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}

// Scan implements sql.Scanner for database retrieval.
func (d *Decimal) Scan(value interface{}) error {
	if value == nil {
		d.value = big.NewRat(0, 1)
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan type %T into Decimal", value)
	}

	dec, err := NewDecimal(str)
	if err != nil {
		return err
	}

	d.value = dec.value
	return nil
}
