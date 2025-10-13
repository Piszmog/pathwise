package contextkey

type Key string

const (
	KeyCorrelationID Key = "correlation_id"
	KeyUserID        Key = "user_id"
)
