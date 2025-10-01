package tracker

const (
	// This should be used internally to track calls to the vendor
	SubjectCallCompleted = "nayla.PROJECT_NAME.calls.completed"

	// This can be used for external services to get more info about the call after processing it
	SubjectCallTracked = "nayla.PROJECT_NAME.calls.tracked"
)
