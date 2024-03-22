package simulator

import (
	"math"

	"gonum.org/v1/plot/plotter"
)

type WaitingLine struct {
	// State vars
	t_arr 		float64
	t_dep 		float64
	locked 		int64
	working 	int64

	// Cummulative vars
	T_att_cum 	float64
	T_cum 		float64
	T_occ 		float64
	T_locked	float64

	// Final vars
	Taux_locked	float64
	Taux_arr	float64
	Miu_obs		float64
	Taux_occ	float64

	// Pointers
	queue		*Queue
	Previous	*WaitingLine
	Next		*WaitingLine

	q 			int64
	p			int64

	Pgraph		plotter.XYs
	Qgraph		plotter.XYs
	Wgraph		plotter.XYs
	Lgraph		plotter.XYs
	QSgraph		plotter.XYs

	// Calcs
	Generate_tmp_service func()float64
	Generate_tmp_arrivee func()float64
}

func (s WaitingLine) GetFirstEventTime() float64 {
	if s.t_arr < s.t_dep {
		return s.t_arr
	}
	return s.t_dep
}

func (s *WaitingLine) CalcCumulatives(delta float64) {
	s.T_cum += delta*( float64(s.queue.Size) + float64(s.working) + float64(s.locked))
	s.T_att_cum += delta*float64(s.queue.Size)
	s.T_occ += delta*float64(s.working)
	s.T_locked += delta*float64(s.locked)
}

func (s *WaitingLine) updateGraphs(t float64) {
	s.Qgraph = append(s.Qgraph, plotter.XY{X : t, Y: float64(s.q)})
	s.Pgraph = append(s.Pgraph, plotter.XY{X : t, Y: float64(s.p)})
	s.Wgraph	= append(s.Wgraph, plotter.XY{X : t, Y: float64(s.working)})
	s.Lgraph	= append(s.Lgraph, plotter.XY{X : t, Y: float64(s.locked)})
	s.QSgraph	= append(s.QSgraph, plotter.XY{X : t, Y: float64(s.queue.Size)})
}

func (s *WaitingLine) CalcFinals(t_max float64) {
	s.Taux_arr = float64(s.q) / t_max
	s.Taux_locked = 100*s.T_locked/t_max
	s.Miu_obs = float64(s.p) / s.T_occ
	s.Taux_occ = 100*s.T_occ/t_max
}

func (s *WaitingLine) SetFirstArrival(t float64) {
	s.t_arr = t
}

func (s *WaitingLine) SetQueueSize(size int64) {
	s.queue.MaxSize = size
}

func NewWaitingLine() WaitingLine {
	return WaitingLine{
		t_arr: 0,
		t_dep: math.MaxInt64,
		locked: 0,
		working: 0,
		T_att_cum: 0,
		T_cum : 0,
		T_occ : 0,
		T_locked : 0,
		
		queue: &Queue{
			Size : 0,
			MaxSize: QUEUE_INF,
		},
		Next: nil,
		Previous: nil,

		Generate_tmp_service: NoArr,
		Generate_tmp_arrivee: NoArr,
	}
}

func NoArr ()float64 {
	return math.MaxFloat64
}