package collectors

func roundTo1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func roundSlice(s []float64) []float64 {
	out := make([]float64, len(s))
	for i, v := range s {
		out[i] = roundTo1(v)
	}
	return out
}
