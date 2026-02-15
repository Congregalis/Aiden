package telegram

import "sync/atomic"

type PollingMetrics struct {
	successCount atomic.Uint64
	failureCount atomic.Uint64
}

func (m *PollingMetrics) RecordSuccess() (uint64, uint64) {
	success := m.successCount.Add(1)
	failure := m.failureCount.Load()
	return success, failure
}

func (m *PollingMetrics) RecordFailure() (uint64, uint64) {
	failure := m.failureCount.Add(1)
	success := m.successCount.Load()
	return success, failure
}

func (m *PollingMetrics) Snapshot() (uint64, uint64) {
	return m.successCount.Load(), m.failureCount.Load()
}
