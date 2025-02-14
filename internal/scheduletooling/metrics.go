package scheduletooling

type MetricsCollector interface {
	JobStarted(name string) (finished func(error))
	SkippedJob(name string)
}
