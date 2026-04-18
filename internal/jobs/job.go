package jobs

type Status string

const (
	Pending   Status = "pending"
	Running   Status = "running"
	Failed    Status = "failed"
	Succeeded Status = "succeeded"
)
