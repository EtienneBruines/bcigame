package systems

import "log"

func Center(numbers []float64) {
	mean := Mean(numbers)
	for numIndex := range numbers {
		numbers[numIndex] -= mean
	}
}

func Mean(numbers []float64) float64 {
	total := 0.0
	for _, num := range numbers {
		total += num
	}
	return total / float64(len(numbers))
}

func Detrend(numbers []float64) {
	sumX := len(numbers) * len(numbers) / 2
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0
	for X, Y := range numbers {
		sumY += Y
		sumXY += float64(X) * Y
		sumX2 += float64(X * X)
	}
	log.Println(sumX)
}
