package session

import "testing"

func BenchmarkManagerNewClientID(b *testing.B) {
	m := NewManager()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.NewClientID()
	}
}
