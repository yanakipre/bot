package worker

import "github.com/google/uuid"

type JobResult []RunResult

type RunResult struct {
	AccountID    *string
	UserID       *uuid.UUID
	ProjectID    *string
	BranchID     *string
	EndpointID   *string
	PeriodID     *uuid.UUID
	OperationID  *uuid.UUID
	PageserverID *int64
	Error        error
}
