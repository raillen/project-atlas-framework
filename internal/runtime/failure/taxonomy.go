package failure

type Class string

const (
	Validation          Class = "validation"
	ApprovalBlocked     Class = "approval_blocked"
	ResourceUnavailable Class = "resource_unavailable"
	RateLimited         Class = "rate_limited"
	ProviderTransient   Class = "provider_transient"
	ProviderPermanent   Class = "provider_permanent"
	ToolFailure         Class = "tool_failure"
	Timeout             Class = "timeout"
	BudgetExhausted     Class = "budget_exhausted"
	ContextOverflow     Class = "context_overflow"
	SideEffectUnknown   Class = "side_effect_unknown"
	VerificationFailure Class = "verification_failure"
	InternalInvariant   Class = "internal_invariant"
	Cancelled           Class = "cancelled"
)

func (c Class) Retryable() bool {
	return c == RateLimited || c == ProviderTransient || c == Timeout || c == ResourceUnavailable
}
