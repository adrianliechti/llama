package cohere

func convertError(err error) error {
	return err
}

func toFloat32(input []float64) []float32 {
	result := make([]float32, len(input))

	for i, v := range input {
		result[i] = float32(v)
	}

	return result
}
