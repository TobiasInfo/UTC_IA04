package utils

// FIFOQueue represents a simple FIFO queue
type FIFOQueue[T comparable] struct {
	items []T
}

// Enqueue adds an item to the queue
func (q *FIFOQueue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

// Enqueue if not prensent for now on the queue
func (q *FIFOQueue[T]) EnqueueIfNotPresent(item T) {
	for _, i := range q.items {
		if i == item {
			return
		}
	}
	q.items = append(q.items, item)
}

// Dequeue removes and returns the first item from the queue
func (q *FIFOQueue[T]) Dequeue() (T, bool) {
	var zero T
	if len(q.items) == 0 {
		return zero, false // Return false if the queue is empty
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

// Size returns the number of items in the queue
func (q *FIFOQueue[T]) Size() int {
	return len(q.items)
}

// // ExampleClass has a FIFOQueue as an attribute
// type ExampleClass struct {
// 	queue FIFOQueue
// }

// // AddToQueue adds an item to the class's queue
// func (e *ExampleClass) AddToQueue(item int) {
// 	e.queue.Enqueue(item)
// }

// // ProcessQueue processes the first item in the queue
// func (e *ExampleClass) ProcessQueue() {
// 	item, ok := e.queue.Dequeue()
// 	if ok {
// 		fmt.Printf("Processed item: %d\n", item)
// 	} else {
// 		fmt.Println("Queue is empty, nothing to process.")
// 	}
// }
