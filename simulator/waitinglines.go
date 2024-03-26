package simulator

import (
	"math"

	"gonum.org/v1/plot/plotter"
)

type WaitingLine struct {
	// Config vars
	n_serv int

	// State vars
	t_arr 		float64
	t_dep 		[]float64
	locked 		[]int64
	working 	[]int64
	

	// Cummulative vars
	T_att_cum 	float64
	T_cum 		float64
	T_occ 		[]float64
	T_locked	[]float64

	// Final vars
	Taux_locked	[]float64
	Taux_arr	float64
	Miu_obs		[]float64
	Miu_wl_obs	float64
	Taux_occ	[]float64
	Taux_wl_occ float64

	// Pointers
	queue		*Queue
	Previous	*WaitingLine
	Next		*WaitingLine

	q 			int64
	p			[]int64

	Pgraph		plotter.XYs
	Qgraph		plotter.XYs
	Wgraph		[]plotter.XYs
	Lgraph		[]plotter.XYs
	QSgraph		plotter.XYs

	// Calcs
	generate_tmp_service []func()float64
	generate_tmp_arrivee func()float64
}

func (s WaitingLine) GetFirstDeparture() float64 {
	t := s.t_dep[0]
	for i := 0 ; i < s.n_serv ; i++ {
		if t > s.t_dep[i] {
			t = s.t_dep[i]
		}
	}
	return t
}

func (s WaitingLine) GetFirstEventTime() float64 {
	t_dep := s.GetFirstDeparture()
	if s.t_arr < t_dep {
		return s.t_arr
	}
	return t_dep
}

func sum(s []float64) float64 {
	t := float64(0)
	for _,item := range s {
		t += item
	}
	return t
}

func sum_i(s []int64) int64 {
	t := int64(0)
	for _,item := range s {
		t += item
	}
	return t
}

func (s *WaitingLine) CalcCumulatives(delta float64) {
	s.T_att_cum += delta*float64(s.queue.Size)
	for i := 0 ; i < s.n_serv ; i++ {
		s.T_occ[i] += delta*float64(s.working[i])
		s.T_locked[i] += delta*float64(s.locked[i])
	}
	s.T_cum = s.T_att_cum + sum(s.T_occ) + sum(s.T_locked)
}

func (s *WaitingLine) updateGraphs(t float64) {
	s.Qgraph = append(s.Qgraph, plotter.XY{X : t, Y: float64(s.q)})
	s.Pgraph = append(s.Pgraph, plotter.XY{X : t, Y: float64(sum_i(s.p))})
	for i := 0 ; i < s.n_serv ; i++ {
		s.Wgraph[i]	= append(s.Wgraph[i], plotter.XY{X : t, Y: float64(s.working[i])})
		s.Lgraph[i]	= append(s.Lgraph[i], plotter.XY{X : t, Y: float64(s.locked[i])})
	}
	s.QSgraph	= append(s.QSgraph, plotter.XY{X : t, Y: float64(s.queue.Size)})
}

func (s *WaitingLine) CalcFinals(t_max float64) {
	s.Taux_arr = float64(s.q) / t_max
	for i := 0 ; i < s.n_serv ; i++ {
		s.Taux_locked[i] = 100*s.T_locked[i]/t_max
		s.Miu_obs[i] = float64(s.p[i]) / s.T_occ[i]
		s.Taux_occ[i] = 100*s.T_occ[i]/t_max
	}
	s.Miu_wl_obs = sum(s.Miu_obs)
	s.Taux_wl_occ = sum(s.Taux_occ)/float64(s.n_serv)
}

func (s *WaitingLine) SetFirstArrival(t float64) {
	s.t_arr = t
}

func (s *WaitingLine) SetQueueSize(size int64) {
	s.queue.MaxSize = size
}

func (s WaitingLine) getFistServerNotBusy() int {
	for i := 0 ; i < s.n_serv ; i++ {
		if s.working[i] == 0 && s.locked[i] == 0 {
			return i
		}
	}
	return -1
}

func (s WaitingLine) getFistServerLocked() int {
	for i := 0 ; i < s.n_serv ; i++ {
		if s.locked[i] == 1 {
			return i
		}
	}
	return -1
}

func (s WaitingLine) getCurrentDeparturingServer(t float64) int {
	for i := 0 ; i < s.n_serv ; i++ {
		if s.t_dep[i] == t {
			return i
		}
	}
	return -1
} 

func (s *WaitingLine) SetArrivalTimeGenerator(f func()float64) {
	s.generate_tmp_arrivee = f
}

func (s *WaitingLine) SetServiceTimeGenerator(f func()float64, sid int) {
	if (sid < s.n_serv) {
		s.generate_tmp_service[sid] = f
	}
}

func NewWaitingLine(server_number int) WaitingLine {
	x :=  WaitingLine{
		n_serv: server_number,
		t_arr: 0,
		t_dep: make([]float64,server_number), // TODO PUT TO MAX
		locked: make([]int64,server_number),
		working: make([]int64,server_number),
		T_att_cum: 0,
		T_cum : 0,
		T_occ : make([]float64,server_number),
		T_locked : make([]float64,server_number),
		Taux_locked: make([]float64,server_number),
		Miu_obs: make([]float64,server_number),
		Taux_occ : make([]float64,server_number),
		Wgraph: make([]plotter.XYs, server_number),
		Lgraph: make([]plotter.XYs, server_number),

		p : make([]int64, server_number),

		
		queue: &Queue{
			Size : 0,
			MaxSize: QUEUE_INF,
		},
		Next: nil,
		Previous: nil,

		generate_tmp_service: make([]func()float64,server_number),
		generate_tmp_arrivee: NoArr,
	}

	return x
}

func NoArr ()float64 {
	return math.MaxFloat64
}