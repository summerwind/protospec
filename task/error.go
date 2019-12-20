package task

type Failed struct {
	Requirement string
	Expect      []string
	Actual      string
}

func (e *Failed) Error() string {
	return "failed"
}

type Skipped struct {
	Reason string
}

func (e *Skipped) Error() string {
	return "skipped"
}
