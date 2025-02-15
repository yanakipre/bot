package codeerr

type ErrorCode string

const (
	ProjectsLimitExceeded               ErrorCode = "PROJECTS_LIMIT_EXCEEDED"
	BranchesLimitExceeded               ErrorCode = "BRANCHES_LIMIT_EXCEEDED"
	BranchesProtectedLimitExceeded      ErrorCode = "BRANCHES_PROTECTED_LIMIT_EXCEEDED"
	RootBranchesLimitExceeded           ErrorCode = "ROOT_BRANCHES_LIMIT_EXCEEDED"
	ActiveEndpointsLimitExceeded        ErrorCode = "ACTIVE_ENDPOINTS_LIMIT_EXCEEDED"
	EndpointsLimitExceeded              ErrorCode = "ENDPOINTS_LIMIT_EXCEEDED"
	VercelEnvVarsExist                  ErrorCode = "VERCEL_ENV_VARS_EXIST"
	BranchSizeLimitExceeded             ErrorCode = "BRANCH_SIZE_LIMIT_EXCEEDED_THRESHOLD"
	NonDefaultBranchComputeTimeExceeded ErrorCode = "NON_PRIMARY_BRANCH_COMPUTE_TIME_EXCEEDED"
	ComputeTimeQuotaExceeded            ErrorCode = "COMPUTE_TIME_EXCEEDED"
	DataTransferLimitExceeded           ErrorCode = "DATA_TRANSFER_LIMIT_EXCEEDED"
	ProjectIDAlreadyExists              ErrorCode = "PROJECT_ID_ALREADY_EXISTS"
	RolesLimitExceeded                  ErrorCode = "ROLES_LIMIT_EXCEEDED"
	DatabasesLimitExceeded              ErrorCode = "DATABASES_LIMIT_EXCEEDED"
	VPCEndpointsLimitExceeded           ErrorCode = "VPC_ENDPOINTS_LIMIT_EXCEEDED"
	VPCEndpointRemoved                  ErrorCode = "VPC_ENDPOINT_REMOVED"
	VPCEndpointAccountOwnershipError    ErrorCode = "VPC_ENDPOINT_ACCOUNT_OWNERSHIP_ERROR"
)

func Wrap(code ErrorCode, err error) *Error {
	return &Error{code: code, err: err}
}
