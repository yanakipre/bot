package status

import "github.com/yanakipre/bot/internal/status/details/reason"

// EnrichWithUserFacingMessage adds a user-facing message to the Status, inspecting its Reason.
// If the UserFacingMessage is already set, the method will not overwrite it.
// If your presenter layer needs some more specific user-facing messages, set them before calling this method.
func (s *Status) EnrichWithUserFacingMessage() {
	dt := s.Details()

	if dt.UserFacingMessage != nil {
		// UserFacingMessage is already set, don't overwrite it.
		return
	}

	if dt.ErrorInfo == nil {
		// We don't have eny ErrorInfo to enrich the message with.
		return
	}

	switch dt.ErrorInfo.Reason {
	case reason.RoleProtected:
		s.WithUserFacingMessage(
			"The role you're attempting to use is protected and cannot be used for password-based authentication. " +
				"Please pick a different role and try again.",
		)
	case reason.ResourceNotFound:
		// Fallback, usually a more specific resource error should be present, e.g., ProjectNotFound.
		s.WithUserFacingMessage(
			"The requested resource could not be found, or you don't have access to it. " +
				"Please check the provided ID and try again.",
		)
	case reason.ProjectNotFound:
		s.WithUserFacingMessage(
			"The requested project could not be found, or you don't have access to it. " +
				"Please check the provided ID and try again.",
		)
	case reason.EndpointNotFound:
		s.WithUserFacingMessage(
			"The requested endpoint could not be found, or you don't have access to it. " +
				"Please check the provided ID and try again.",
		)
	case reason.BranchNotFound:
		s.WithUserFacingMessage(
			"The requested branch could not be found, or you don't have access to it. " +
				"Please check the provided ID and try again.",
		)
	case reason.RateLimitExceeded:
		s.WithUserFacingMessage(
			"You've exceeded the rate limit. Please wait a moment and try again.",
		)
	case reason.NonDefaultBranchComputeTimeExceeded:
		s.WithUserFacingMessage(
			"You've exceeded the compute time for this branch. " +
				"Upgrade your plan to increase limits or connect to another branch.",
		)
	case reason.ActiveTimeQuotaExceeded:
		s.WithUserFacingMessage(
			"Your project has exceeded the active time quota. Upgrade your plan to increase limits.",
		)
	case reason.ComputeTimeQuotaExceeded:
		s.WithUserFacingMessage(
			"Your account or project has exceeded the compute time quota. Upgrade your plan to increase limits.",
		)
	case reason.WrittenDataQuotaExceeded:
		s.WithUserFacingMessage(
			"Your project has exceeded the written data quota. Upgrade your plan to increase limits.",
		)
	case reason.DataTransferQuotaExceeded:
		s.WithUserFacingMessage(
			"Your project has exceeded the data transfer quota. Upgrade your plan to increase limits.",
		)
	case reason.LogicalSizeQuotaExceeded:
		s.WithUserFacingMessage(
			"Your project has exceeded the logical size quota. Upgrade your plan to increase limits.",
		)
	case reason.RunningOperations:
		s.WithUserFacingMessage(
			"Your project already has running operations. Please wait until they complete, and then try again.",
		)
	case reason.ActiveEndpointsLimitExceeded:
		s.WithUserFacingMessage(
			"You have exceeded the limit of concurrently active endpoints. Please suspend some endpoints and try again.",
		)
	}
}
