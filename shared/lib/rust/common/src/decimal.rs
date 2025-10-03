use rust_decimal::Decimal as RustDecimal;
use serde::{Deserialize, Serialize};
use std::fmt;
use std::ops::{Add, Div, Mul, Sub};
use std::str::FromStr;

/// High-precision decimal type for financial calculations.
/// Wrapper around rust_decimal to avoid floating point errors.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub struct Decimal(RustDecimal);

impl Decimal {
    /// Creates a new Decimal from a string.
    pub fn from_str(s: &str) -> Result<Self, rust_decimal::Error> {
        Ok(Decimal(RustDecimal::from_str(s)?))
    }

    /// Creates a new Decimal from an i64.
    pub fn from_i64(value: i64) -> Self {
        Decimal(RustDecimal::from(value))
    }

    /// Creates a new Decimal from an f64.
    /// Note: May introduce floating point errors. Use from_str for exact values.
    pub fn from_f64(value: f64) -> Option<Self> {
        RustDecimal::from_f64_retain(value).map(Decimal)
    }

    /// Returns a decimal representing zero.
    pub fn zero() -> Self {
        Decimal(RustDecimal::ZERO)
    }

    /// Returns a decimal representing one.
    pub fn one() -> Self {
        Decimal(RustDecimal::ONE)
    }

    /// Returns the float64 representation (may lose precision).
    pub fn to_f64(&self) -> Option<f64> {
        self.0.to_f64()
    }

    /// Returns the string representation with full precision.
    pub fn to_string(&self) -> String {
        self.0.to_string()
    }

    /// Returns true if the decimal is zero.
    pub fn is_zero(&self) -> bool {
        self.0.is_zero()
    }

    /// Returns true if the decimal is positive.
    pub fn is_positive(&self) -> bool {
        self.0.is_sign_positive()
    }

    /// Returns true if the decimal is negative.
    pub fn is_negative(&self) -> bool {
        self.0.is_sign_negative()
    }

    /// Returns the absolute value.
    pub fn abs(&self) -> Self {
        Decimal(self.0.abs())
    }

    /// Returns the negation.
    pub fn neg(&self) -> Self {
        Decimal(-self.0)
    }

    /// Rounds to the specified number of decimal places.
    pub fn round_dp(&self, dp: u32) -> Self {
        Decimal(self.0.round_dp(dp))
    }

    /// Truncates to the specified number of decimal places.
    pub fn trunc_dp(&self, dp: u32) -> Self {
        Decimal(self.0.trunc_with_scale(dp))
    }

    /// Returns the minimum of two decimals.
    pub fn min(&self, other: Self) -> Self {
        Decimal(self.0.min(other.0))
    }

    /// Returns the maximum of two decimals.
    pub fn max(&self, other: Self) -> Self {
        Decimal(self.0.max(other.0))
    }

    /// Clamps the decimal to the given range.
    pub fn clamp(&self, min: Self, max: Self) -> Self {
        Decimal(self.0.clamp(min.0, max.0))
    }
}

impl Add for Decimal {
    type Output = Self;

    fn add(self, other: Self) -> Self {
        Decimal(self.0 + other.0)
    }
}

impl Sub for Decimal {
    type Output = Self;

    fn sub(self, other: Self) -> Self {
        Decimal(self.0 - other.0)
    }
}

impl Mul for Decimal {
    type Output = Self;

    fn mul(self, other: Self) -> Self {
        Decimal(self.0 * other.0)
    }
}

impl Div for Decimal {
    type Output = Self;

    fn div(self, other: Self) -> Self {
        Decimal(self.0 / other.0)
    }
}

impl fmt::Display for Decimal {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl From<i64> for Decimal {
    fn from(value: i64) -> Self {
        Decimal::from_i64(value)
    }
}

impl From<RustDecimal> for Decimal {
    fn from(value: RustDecimal) -> Self {
        Decimal(value)
    }
}

impl From<Decimal> for RustDecimal {
    fn from(value: Decimal) -> Self {
        value.0
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_decimal_operations() {
        let a = Decimal::from_i64(10);
        let b = Decimal::from_i64(5);

        assert_eq!(a + b, Decimal::from_i64(15));
        assert_eq!(a - b, Decimal::from_i64(5));
        assert_eq!(a * b, Decimal::from_i64(50));
        assert_eq!(a / b, Decimal::from_i64(2));
    }

    #[test]
    fn test_decimal_precision() {
        let a = Decimal::from_str("0.1").unwrap();
        let b = Decimal::from_str("0.2").unwrap();
        let expected = Decimal::from_str("0.3").unwrap();

        assert_eq!(a + b, expected);
    }
}
