package audit

type NoopLogger struct{}

func (NoopLogger) LogEvent(_ Event) {}

func (NoopLogger) ListEvents(_ EventQuery) []Event {
	return []Event{}
}

func (NoopLogger) CountEvents(_ EventQuery) (int64, bool) {
	return 0, false
}
