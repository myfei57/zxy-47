package env

type TemperatureFilter struct {
	window int
}

func NewTemperatureFilter(window int) *TemperatureFilter {
	if window < 2 {
		window = 2
	}
	return &TemperatureFilter{window: window}
}

func (f *TemperatureFilter) Average(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range samples {
		total += value
	}
	return total / float64(len(samples))
}

func (f *TemperatureFilter) Smooth(history []float64, sample float64) []float64 {
	next := append(append([]float64{}, history...), sample)
	if len(next) > f.window {
		next = next[len(next)-f.window:]
	}
	return next
}

func (f *TemperatureFilter) LastAverage(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	window := samples
	if len(window) > f.window {
		window = window[len(window)-f.window:]
	}
	return f.Average(window)
}

func (f *TemperatureFilter) Window() int {
	return f.window
}

func RawMax(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	max := samples[0]
	for _, value := range samples[1:] {
		if value > max {
			max = value
		}
	}
	return max
}
