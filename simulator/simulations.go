package simulator

import (
	"fmt"
	"strconv"

	"gonum.org/v1/plot/plotter"
)

type Simulation struct {
	WaitingLines		[]*WaitingLine
	
	T			float64
	T_max		float64

	Q			int64
	P			int64

	Qgraph			plotter.XYs
	Pgraph			plotter.XYs

	T_moy		float64
	T_att_moy	float64
	N_moy		float64
	Taux_arr	float64
	Taux_occ	float64
	Miu_obs		float64

	T_cum		float64
	T_att_cum	float64
	T_occ		float64
}

func (s *Simulation) TestRandomFunctions() {
	fmt.Printf("Testing random functions...\n")
	for i,wl := range s.WaitingLines {
		sum := float64(0)
		for j := 0 ; j < 10000 ; j++ {
			sum += wl.Generate_tmp_arrivee()
		}
		fmt.Printf("Tested mean for arr of wl %v : %v\n",i,sum/10000)
		sum = 0
		for j := 0 ; j < 10000 ; j++ {
			sum += wl.Generate_tmp_service()
		}
		fmt.Printf("Tested mean for service of wl %v : %v\n\n",i,sum/10000)
	}
}

func (s *Simulation) updateGraphs() {
	s.Qgraph = append(s.Qgraph, plotter.XY{X : s.T, Y: float64(s.Q)})
	s.Pgraph = append(s.Pgraph, plotter.XY{X : s.T, Y: float64(s.P)})
}

func (s *Simulation) Run() {
	s.ShowHeader()
	for s.GetFirstEventTime() <= s.T_max {
		index := s.GetFirstEventIndex()
		serv := s.WaitingLines[index]
		if serv.t_arr < serv.t_dep { // Arrivée d'une piece
			s.updateGraphs()
			s.Q++
			serv.q++
			delta := serv.t_arr - s.T
			s.T = serv.t_arr

			for _,wl := range s.WaitingLines {
				wl.updateGraphs(s.T)
				wl.CalcCumulatives(delta)
			}


			serv.t_arr = s.T + serv.Generate_tmp_arrivee()

			if serv.working == 0 && serv.locked == 0 {
				serv.working = 1
				serv.t_dep = s.T + serv.Generate_tmp_service()
			} else {
				if serv.queue.Size < serv.queue.MaxSize {
					serv.queue.Size ++
				} else {
					fmt.Printf("====== queue OVERFLOW ======")
					break
				}
			}
			s.ShowStat("ARR"+strconv.Itoa(index))
			s.updateGraphs()
			for _,wl := range s.WaitingLines {
				wl.updateGraphs(s.T)
			}
		} else { // Départ d'une pièce
			s.updateGraphs()
			serv.p++
			delta := serv.t_dep - s.T
			s.T = serv.t_dep

			for _,wl := range s.WaitingLines {
				wl.CalcCumulatives(delta)
				wl.updateGraphs(s.T)
			}

			if serv.Next == nil || serv.Next.queue.Size < serv.Next.queue.MaxSize {
				// Gestion serveur actuel
				if serv.queue.Size > 0 {
					serv.queue.Size --
					serv.t_dep = s.T + serv.Generate_tmp_service()
					if serv.Previous != nil && serv.Previous.locked == 1 {
						if serv.Previous.queue.Size > 0 {
							serv.Previous.queue.Size --
							serv.Previous.working = 1
							serv.Previous.t_dep = s.T + serv.Previous.Generate_tmp_service()
						}
						serv.Previous.locked = 0
						serv.queue.Size ++
						serv.q++
					}
				} else {
					serv.working = 0
					serv.t_dep = s.T_max + 1
				}

				// Gestion Serveur next
				if serv.Next != nil {
					serv.Next.q ++
					if serv.Next.queue.Size == 0 && serv.Next.working == 0 && serv.Next.locked == 0 {
						serv.Next.working = 1
						serv.Next.t_dep = s.T + serv.Next.Generate_tmp_service()
					} else {
						serv.Next.queue.Size ++
					}
				}
			} else if(serv.Next != nil) {
				serv.locked = 1
				serv.working = 0
				serv.t_dep = s.T_max + 1
			} 
			if (serv.Next == nil) {
				s.P++
			}
			s.ShowStat("DEP"+strconv.Itoa(index))
			s.updateGraphs()
			for _,wl := range s.WaitingLines {
				wl.updateGraphs(s.T)
			}
		}
	}

	delta := s.T_max - s.T
	//s.T = s.T_max
	for _,wl := range s.WaitingLines {
		wl.CalcCumulatives(delta)
	}

	s.CalcAndDisplayResult();
}

func (s Simulation) CalcAndDisplayResult() {
	s.T_cum = 0
	s.T_att_cum = 0
	s.T_occ = 0
	for _,wl := range s.WaitingLines {
		wl.CalcFinals(s.T_max)
		s.T_cum += wl.T_cum
		s.T_att_cum += wl.T_att_cum
		s.T_occ += wl.T_occ
	}

	s.T_moy = s.T_cum/float64(s.Q)
	s.T_att_moy = s.T_att_cum/float64(s.Q)
	s.N_moy = s.T_cum/s.T_max
	s.Taux_arr = float64(s.Q)/s.T_max

	fmt.Printf("\n# ------------------------------------\n# Resultats de FIN de simulation\n# ------------------------------------\n");
    fmt.Printf("# Periode de simulation : %.2f minutes\n", s.T_max);
    fmt.Printf("# Nombre de pièces arrivees : %d\n", s.Q);
    fmt.Printf("# Nombre de pièces produites : %d\n", s.P);
	fmt.Printf("# [temps du dernier evenement simule avant t_max : %.2f]\n", s.T);
	fmt.Printf("# Nombre moyen de pieces dans le systeme : \t%f\n", s.N_moy);
    fmt.Printf("# Temps moyen de sejour passe par une piece : \t%f\n", s.T_moy);
    fmt.Printf("# Temps moyen passe par une piece en attente : \t%f\n", s.T_att_moy);
	
	for i,wl := range s.WaitingLines {
		fmt.Println("#")
		fmt.Printf("# λ_%v (taux d'arrivee de pieces de %v) observe :  %f\n", i,i,wl.Taux_arr);
		fmt.Printf("# Nombre de pièces en attente dans la file %d : %d\n", i, wl.queue.Size);
		fmt.Printf("# Etat du serveur %d à t_max : %d (bloqué : %d)\n", i, wl.working,wl.locked);
		fmt.Printf("# Pourcentage d'occupation du serveur %d : \t%.2f %%\n", i, wl.Taux_occ);
		fmt.Printf("# Pourcentage de bloquage du serveur %d : \t%3.2f %%\n", i, wl.Taux_locked);
		fmt.Printf("# µ_%d (taux de service du serveur %d) observe : %f\n", i, i, wl.Miu_obs);
	}
}

func (s Simulation) GetFirstEventTime() float64 {
	min := s.T_max+1
	for _,wl :=  range s.WaitingLines {
		if min > wl.GetFirstEventTime() {
			min = wl.GetFirstEventTime()
		}
	}
	return min
}

func (s Simulation) GetFirstEventIndex() int {
	min := s.T_max+1
	index := -1
	for i,wl :=  range s.WaitingLines {
		if min > wl.GetFirstEventTime() {
			min = wl.GetFirstEventTime()
			index = i
		}
	}
	return index
}

func (s Simulation) ShowHeader() {
	fmt.Printf("info\tclock\t|\t")
	for index := range s.WaitingLines {
		fmt.Printf("w_%v\tlck_%v\tQs_%v\tq_%v\tp_%v\t|\t",index,index,index,index,index)
	}
	fmt.Printf("q\tp\t|\n")
	fmt.Print("-----------------")
	for i := 0 ; i < len(s.WaitingLines) ; i++ {
		fmt.Print("------------------------------------------------")
	}
	fmt.Print("------------------------\n")
}

func (s Simulation) ShowStat(msg string) {
	fmt.Printf("%v\t%.2f\t|\t",msg,s.T)
	for _,wl := range s.WaitingLines {
		fmt.Printf("%v\t%v\t%v\t%v\t%v\t|\t",wl.working,wl.locked,wl.queue.Size,wl.q,wl.p)
	}
	fmt.Printf("%v\t%v\t|\n",s.Q,s.P)
}

func NewSimulation(t_max float64,WaitingLines ...*WaitingLine) Simulation {
	WaitingLine_list := []*WaitingLine{}

	for i,wl := range WaitingLines {
		if i != 0 {
			wl.Previous = WaitingLines[i-1]
		}
		if i < len(WaitingLines)-1 {
			wl.Next = WaitingLines[i+1]
		}
		WaitingLine_list = append(WaitingLine_list, wl)
	}

	return Simulation{
		WaitingLines: WaitingLine_list,
		Qgraph: make(plotter.XYs,0),
		Pgraph: make(plotter.XYs,0),

		T_max: t_max,
		Q : 0,
		P : 0,
	}
}