package scheduletooling

type MetricsCollector interface {
	JobStarted(name string) (finished func())
	SkippedJob(name string)
}
