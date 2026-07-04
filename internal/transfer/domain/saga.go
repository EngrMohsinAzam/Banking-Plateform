package domain

// SagaState is the lifecycle of a money-movement saga.
type SagaState string

const (
	SagaStateStarted      SagaState = "STARTED"
	SagaStateFraudOK      SagaState = "FRAUD_OK"
	SagaStateComplianceOK SagaState = "COMPLIANCE_OK"
	SagaStatePosted       SagaState = "POSTED"
	SagaStateSettling     SagaState = "SETTLING"
	SagaStateCompleted    SagaState = "COMPLETED"
	SagaStateCompensating SagaState = "COMPENSATING"
	SagaStateCompensated  SagaState = "COMPENSATED"
	SagaStateFailed       SagaState = "FAILED"
)

// SettlementStatus tracks mock sarie settlement.
type SettlementStatus string

const (
	SettlementPending SettlementStatus = "PENDING"
	SettlementSettled SettlementStatus = "SETTLED"
	SettlementFailed  SettlementStatus = "FAILED"
)
