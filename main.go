package main

import (
	"fmt"
	"math"
	"math/rand"
	"symsys/random"
	"symsys/simulator"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func main() {
	s1 := simulator.NewWaitingLine()
	s1.Generate_tmp_service = service1
	s1.Generate_tmp_arrivee = arr1
	s1.SetQueueSize(simulator.QUEUE_INF)

	s2 := simulator.NewWaitingLine()
	s2.Generate_tmp_service = service2
	s2.SetQueueSize(5)
	s2.SetFirstArrival(math.MaxFloat64)

	sim := simulator.NewSimulation(100,&s1,&s2)

	sim.Run()
	fmt.Println()
	sim.TestRandomFunctions()

	p := plot.New()

	err := plotutil.AddLines(
		p,
		"W_0",s1.Wgraph,
		"W_1",s2.Wgraph,
		"QS_0",s1.QSgraph,
		"QS_1",s2.QSgraph,
		"L_0", s1.Lgraph,
	)

	if err != nil {
		panic(err)
	}

	if err := p.Save(12*vg.Inch, 6*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}

func service1() float64{
	return	random.UniformRand(0.5,1.5)
}

func service2() float64 {
	r := rand.Float64()*100
	if r < 10 {
		return 1.5;
	} else if r < 35 {
		return 2;
	} else if r < 75 {
		return 2.5;
	} else if r < 90 {
		return 3;
	} else {
		return 3.5;
	}
}

func arr1() float64 {
	return random.UniformRand(1/0.6 - 0.5, 1/0.6 + 0.5)
}