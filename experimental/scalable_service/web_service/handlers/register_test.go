package handlers

import "testing"

func BenchmarkLoadCfg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadCfg("../html/serviceCfg.yaml")
	}
}
