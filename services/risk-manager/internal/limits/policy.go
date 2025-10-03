package limits

import (
	"fmt"
	"time"
)

// PolicyType defines the severity of a policy
type PolicyType string

const (
	PolicyTypeHard      PolicyType = "hard"      // Block order
	PolicyTypeSoft      PolicyType = "soft"      // Warn only
	PolicyTypeEmergency PolicyType = "emergency" // Trigger emergency stop
)

// PolicyScope defines what the policy applies to
type PolicyScope string

const (
	PolicyScopeAccount  PolicyScope = "account"
	PolicyScopeSymbol   PolicyScope = "symbol"
	PolicyScopeStrategy PolicyScope = "strategy"
)

// PolicyOperator defines comparison operators
type PolicyOperator string

const (
	OperatorLessThan           PolicyOperator = "less_than"
	OperatorLessThanOrEqual    PolicyOperator = "less_than_or_equal"
	OperatorGreaterThan        PolicyOperator = "greater_than"
	OperatorGreaterThanOrEqual PolicyOperator = "greater_than_or_equal"
	OperatorEqual              PolicyOperator = "equal"
	OperatorNotEqual           PolicyOperator = "not_equal"
)

// Policy represents a risk policy
type Policy struct {
	ID        string
	Name      string
	Type      PolicyType
	Metric    string
	Operator  PolicyOperator
	Threshold float64
	Scope     PolicyScope
	ScopeID   string // Symbol name, strategy ID, etc.
	Action    string
	Enabled   bool
	Priority  int
	Metadata  map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int
}

// PolicyViolation represents a violated policy
type PolicyViolation struct {
	Policy       *Policy
	MetricValue  float64
	ThresholdValue float64
	Message      string
	Timestamp    time.Time
}

// PolicyEngine evaluates policies against metrics
type PolicyEngine struct {
	policies []*Policy
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		policies: make([]*Policy, 0),
	}
}

// LoadPolicies loads policies into the engine
func (e *PolicyEngine) LoadPolicies(policies []*Policy) {
	e.policies = policies
}

// GetPolicies returns all loaded policies
func (e *PolicyEngine) GetPolicies() []*Policy {
	return e.policies
}

// GetApplicablePolicies returns policies applicable to the given context
func (e *PolicyEngine) GetApplicablePolicies(symbol, strategyID string) []*Policy {
	applicable := make([]*Policy, 0)

	for _, policy := range e.policies {
		if !policy.Enabled {
			continue
		}

		// Check scope
		switch policy.Scope {
		case PolicyScopeAccount:
			// Always applicable
			applicable = append(applicable, policy)

		case PolicyScopeSymbol:
			// Check if symbol matches
			if policy.ScopeID == "" || policy.ScopeID == symbol {
				applicable = append(applicable, policy)
			}

		case PolicyScopeStrategy:
			// Check if strategy matches
			if policy.ScopeID == "" || policy.ScopeID == strategyID {
				applicable = append(applicable, policy)
			}
		}
	}

	return applicable
}

// EvaluatePolicy evaluates a single policy against metrics
func (e *PolicyEngine) EvaluatePolicy(policy *Policy, metrics map[string]float64) *PolicyViolation {
	metricValue, exists := metrics[policy.Metric]
	if !exists {
		// Metric not available, skip evaluation
		return nil
	}

	violated := false
	switch policy.Operator {
	case OperatorLessThan:
		violated = metricValue < policy.Threshold
	case OperatorLessThanOrEqual:
		violated = metricValue <= policy.Threshold
	case OperatorGreaterThan:
		violated = metricValue > policy.Threshold
	case OperatorGreaterThanOrEqual:
		violated = metricValue >= policy.Threshold
	case OperatorEqual:
		violated = metricValue == policy.Threshold
	case OperatorNotEqual:
		violated = metricValue != policy.Threshold
	}

	if violated {
		return &PolicyViolation{
			Policy:         policy,
			MetricValue:    metricValue,
			ThresholdValue: policy.Threshold,
			Message: fmt.Sprintf(
				"Policy '%s' violated: %s (%.4f) %s %.4f",
				policy.Name,
				policy.Metric,
				metricValue,
				policy.Operator,
				policy.Threshold,
			),
			Timestamp: time.Now(),
		}
	}

	return nil
}

// EvaluateAll evaluates all applicable policies
func (e *PolicyEngine) EvaluateAll(metrics map[string]float64, symbol, strategyID string) []*PolicyViolation {
	applicable := e.GetApplicablePolicies(symbol, strategyID)
	violations := make([]*PolicyViolation, 0)

	for _, policy := range applicable {
		if violation := e.EvaluatePolicy(policy, metrics); violation != nil {
			violations = append(violations, violation)
		}
	}

	return violations
}

// GetViolationsByType groups violations by policy type
func GetViolationsByType(violations []*PolicyViolation) map[PolicyType][]*PolicyViolation {
	grouped := make(map[PolicyType][]*PolicyViolation)

	for _, v := range violations {
		policyType := v.Policy.Type
		grouped[policyType] = append(grouped[policyType], v)
	}

	return grouped
}

// HasEmergencyViolations checks if any violations are emergency type
func HasEmergencyViolations(violations []*PolicyViolation) bool {
	for _, v := range violations {
		if v.Policy.Type == PolicyTypeEmergency {
			return true
		}
	}
	return false
}

// FormatViolations formats violations into strings
func FormatViolations(violations []*PolicyViolation) []string {
	messages := make([]string, len(violations))
	for i, v := range violations {
		messages[i] = v.Message
	}
	return messages
}

// MetricsFromRiskMetrics converts RiskMetrics to a map for policy evaluation
func MetricsFromRiskMetrics(leverage, marginRatio, drawdownDaily, drawdownMax float64,
	positionConcentration map[string]float64) map[string]float64 {

	metrics := map[string]float64{
		"leverage":        leverage,
		"margin_ratio":    marginRatio,
		"drawdown_daily":  drawdownDaily,
		"drawdown_max":    drawdownMax,
	}

	// Add position concentration metrics
	for symbol, concentration := range positionConcentration {
		key := fmt.Sprintf("concentration_%s", symbol)
		metrics[key] = concentration
	}

	return metrics
}
