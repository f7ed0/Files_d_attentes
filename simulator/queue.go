package simulator

import "math"

const (
	QUEUE_INF = math.MaxInt64
)

type Queue struct {
	MaxSize		int64
	Size		int64
}