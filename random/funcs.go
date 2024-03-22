package random

import (
	"math"
	"math/rand"
)

func UniformRand(min,max float64) float64 {
	return (rand.Float64())*(max-min) + min
}

func PoissonRand(lambda float64) float64 {
	var k int = 0

	p := math.Exp(-lambda)
	F := p

	Z := UniformRand(0,1)

	for F < Z {
		k++
		p *= (lambda/float64(k))
		F += p
	}

	return float64(k)
}


// Return a random value following the law of density of probability f_dens_prob using the rejectMethod.
// a and b are the min and max absisse value of the f_dens_prob and c is its max ordinal value.
func RejectMethod(f_dens_prob func(float64)float64, a,b,M float64) float64 {
	x := UniformRand(a,b)
	y := UniformRand(0,M)
	for f_dens_prob(x) > y {
		x = UniformRand(a,b)
		y = UniformRand(0,M)
	}
	return x
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n*factorial(n-1)
}