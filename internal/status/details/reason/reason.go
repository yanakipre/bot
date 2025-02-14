// Package reason provides well-known reasons for errors.
//
// These reasons are used in ErrorInfo to provide a machine-readable reason for the error.
// By convention, error reasons are in UPPER_SNAKE_CASE.
package reason

type Reason string

const (
	// RoleProtected indicates that the role is protected and the attempted operation is not permitted on protected roles.
	RoleProtected = Reason("ROLE_PROTECTED")

	// ResourceNotFound indicates that a resource (project, endpoint, branch, etc.) wasn't found,
	// usually due to the provided ID not being correct or because the subject doesn't have enough permissions to
	// access the requested resource.
	// Prefer a more specific reason if possible, e.g., ProjectNotFound, EndpointNotFound, etc.
	ResourceNotFound = Reason("RESOURCE_NOT_FOUND")

	// ProjectNotFound indicates that the project wasn't found, usually due to the provided ID not being correct,
	// or that the subject doesn't have enough permissions to access the requested project.
	ProjectNotFound = Reason("PROJECT_NOT_FOUND")

	// EndpointNotFound indicates that the endpoint wasn't found, usually due to the provided ID not being correct,
	// or that the subject doesn't have enough permissions to access the requested endpoint.
	EndpointNotFound = Reason("ENDPOINT_NOT_FOUND")

	// BranchNotFound indicates that the branch wasn't found, usually due to the provided ID not being correct,
	// or that the subject doesn't have enough permissions to access the requested branch.
	BranchNotFound = Reason("BRANCH_NOT_FOUND")

	// RateLimitExceeded indicates that the rate limit for the operation has been exceeded.
	RateLimitExceeded = Reason("RATE_LIMIT_EXCEEDED")

	// NonDefaultBranchComputeTimeExceeded indicates that the compute time quota of non-default branches has been
	// exceeded.
	NonDefaultBranchComputeTimeExceeded = Reason("NON_PRIMARY_BRANCH_COMPUTE_TIME_EXCEEDED")

	// ActiveTimeQuotaExceeded indicates that the active time quota was exceeded.
	ActiveTimeQuotaExceeded = Reason("ACTIVE_TIME_QUOTA_EXCEEDED")

	// ComputeTimeQuotaExceeded indicates that the compute time quota was exceeded.
	ComputeTimeQuotaExceeded = Reason("COMPUTE_TIME_QUOTA_EXCEEDED")

	// WrittenDataQuotaExceeded indicates that the written data quota was exceeded.
	WrittenDataQuotaExceeded = Reason("WRITTEN_DATA_QUOTA_EXCEEDED")

	// DataTransferQuotaExceeded indicates that the data transfer quota was exceeded.
	DataTransferQuotaExceeded = Reason("DATA_TRANSFER_QUOTA_EXCEEDED")

	// LogicalSizeQuotaExceeded indicates that the logical size quota was exceeded.
	LogicalSizeQuotaExceeded = Reason("LOGICAL_SIZE_QUOTA_EXCEEDED")

	// RunningOperations indicates that the project already has some running operations
	// and scheduling of new ones is prohibited.
	RunningOperations = Reason("RUNNING_OPERATIONS")

	// ConcurrencyLimitReached indicates that the concurrency limit for an action was reached.
	ConcurrencyLimitReached = Reason("CONCURRENCY_LIMIT_REACHED")

	// LockAlreadyTaken indicates that we attempted to take a lock that was already taken.
	LockAlreadyTaken = Reason("LOCK_ALREADY_TAKEN")

	// ActiveEndpointsLimitExceeded indicates that the limit of concurrently active endpoints was exceeded.
	ActiveEndpointsLimitExceeded = Reason("ACTIVE_ENDPOINTS_LIMIT_EXCEEDED")
)
