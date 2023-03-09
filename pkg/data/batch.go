package data

// Batch divides the provided slice of things into batches of the specified maximum size.
func Batch[T any](items []T, batchSize int) [][]T {
	batches := make([][]T, 0)
	batch := make([]T, 0)
	for _, item := range items {
		batch = append(batch, item)
		if len(batch) == batchSize {
			batches = append(batches, batch)
			batch = make([]T, 0)
		}
	}
	if len(batch) > 0 {
		batches = append(batches, batch)
	}
	return batches
}
