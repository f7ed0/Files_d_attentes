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

	Qgraph		plotter.XYs
	Pgraph		plotter.XYs

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
			sum += wl.generate_tmp_arrivee()
		}
		fmt.Printf("Tested mean for arr of wl %v : %v (1/E[α] : %v)\n",i,sum/10000,10000/sum)
		for k := 0 ; k < wl.n_serv ; k++ {
			sum = 0
			for j := 0 ; j < 10000 ; j++ {
				sum += wl.generate_tmp_service[k]()
			}
			fmt.Printf("Tested mean for service of server %v wl %v : %v (1/E[β] : %v)\n",k,i,sum/10000,10000/sum)
		}
		fmt.Println()
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
		wl := s.WaitingLines[index]
		if wl.t_arr < wl.GetFirstDeparture() { // Arrivée d'une piece
			s.updateGraphs()
			s.Q++
			wl.q++
			delta := wl.t_arr - s.T
			s.T = wl.t_arr

			for _,wl := range s.WaitingLines {
				wl.updateGraphs(s.T)
				wl.CalcCumulatives(delta)
			}

			wl.t_arr = s.T + wl.generate_tmp_arrivee()

			srv_index := wl.getFistServerNotBusy()
			if srv_index >= 0 {
				wl.working[srv_index] = 1
				wl.t_dep[srv_index] = s.T + wl.generate_tmp_service[srv_index]()
			} else {
				if wl.queue.Size < wl.queue.MaxSize {
					wl.queue.Size ++
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
			srv_index := wl.getCurrentDeparturingServer(wl.GetFirstDeparture())
			wl.p[srv_index]++
			
			delta := wl.t_dep[srv_index] - s.T
			s.T = wl.t_dep[srv_index]

			for _,wl := range s.WaitingLines {
				wl.CalcCumulatives(delta)
				wl.updateGraphs(s.T)
			}

			if wl.Next == nil || wl.Next.queue.Size < wl.Next.queue.MaxSize {
				// Gestion wl actuel
				if wl.queue.Size > 0 {
					wl.queue.Size --
					wl.t_dep[srv_index] = s.T + wl.generate_tmp_service[srv_index]()
					if wl.Previous != nil {
						psid := wl.Previous.getFistServerLocked()
						if psid != 1 {
							if wl.Previous.queue.Size > 0 {
								wl.Previous.queue.Size --
								wl.Previous.working[psid] = 1
								wl.Previous.t_dep[psid] = s.T + wl.Previous.generate_tmp_service[psid]()
							}
							wl.Previous.locked[psid] = 0
							wl.queue.Size ++
							wl.q++
						}
					}
				} else {
					wl.working[srv_index] = 0
					wl.t_dep[srv_index] = s.T_max + 1
				}

				// Gestion wl next
				if wl.Next != nil {
					wl.Next.q ++
					nsid := wl.Next.getFistServerNotBusy()
					if wl.Next.queue.Size == 0 && nsid != 0 {
						wl.Next.working[nsid] = 1
						wl.Next.t_dep[nsid] = s.T + wl.Next.generate_tmp_service[nsid]()
					} else {
						wl.Next.queue.Size ++
					}
				}
			} else if(wl.Next != nil) {
				wl.locked[srv_index] = 1
				wl.working[srv_index] = 0
				wl.t_dep[srv_index] = s.T_max + 1
			} 
			if (wl.Next == nil) {
				s.P++
			}
			s.ShowStat("DEP"+strconv.Itoa(index)+"|"+strconv.Itoa(srv_index))
			s.updateGraphs()
			for _,wl := range s.WaitingLines {
				wl.updateGraphs(s.T)
			}
		}
	}

	delta := s.T_max - s.T
	s.T = s.T_max
	for _,wl := range s.WaitingLines {
		wl.CalcCumulatives(delta)
	}
	s.ShowStat("END")

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
		s.T_occ += sum(wl.T_occ)
	}

	s.T_moy = s.T_cum/float64(s.Q)
	s.T_att_moy = s.T_att_cum/float64(s.Q)
	s.N_moy = s.T_cum/s.T_max
	s.Taux_arr = float64(s.Q)/s.T_max

	fmt.Printf("\n# ---------------------------------------------------------------------------------\n# Resultats de FIN de simulation\n# ---------------------------------------------------------------------------------\n");
    fmt.Printf("# Periode de simulation : \t\t\t\t%.2f minutes\n", s.T_max);
    fmt.Printf("# Nombre de pièces arrivees : \t\t\t\t%d\n", s.Q);
    fmt.Printf("# Nombre de pièces produites : \t\t\t\t%d\n", s.P);
	fmt.Printf("# Temps du dernier evenement simule avant t_max : \t%.2f\n", s.T);
	fmt.Printf("# Nombre moyen de pieces dans le systeme : \t\t%f\n", s.N_moy);
    fmt.Printf("# Temps moyen de sejour passe par une piece : \t\t%f\n", s.T_moy);
    fmt.Printf("# Temps moyen passe par une piece en attente : \t\t%f\n", s.T_att_moy);
	
	for i,wl := range s.WaitingLines {
		fmt.Println("#")
		fmt.Printf("# λ_%v (taux d'arrivee de pieces de %v) observe :  \t%f\n", i,i,wl.Taux_arr);
		fmt.Printf("# Nombre de pièces en attente dans la file %d : \t\t%d\n", i, wl.queue.Size);
		fmt.Printf("# Etat du serveur %d à t_max : \t\t\t\t%d (bloqué : %d)\n", i, wl.working,wl.locked);
		fmt.Printf("# Pourcentage d'occupation du serveur %d : \t\t%.2f %%\n", i, wl.Taux_occ);
		fmt.Printf("# Pourcentage d'occupation global du serveur %d : \t%.2f %%\n", i, wl.Taux_wl_occ);
		fmt.Printf("# Pourcentage de bloquage du serveur %d : \t\t%3.2f %%\n", i, wl.Taux_locked);
		fmt.Printf("# taux de service des serveurs %d observe : \t\t%f\n", i, wl.Miu_obs);
		fmt.Printf("# µ_%d (taux de service global du serveur %d) observe : \t%f\n", i, i, wl.Miu_wl_obs);
	}
	fmt.Println("# ----------------------------------------------------------------------------------")
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
		fmt.Printf("%v\t%v\t%v\t%v\t%v\t|\t",sum_i(wl.working),sum_i(wl.locked),wl.queue.Size,wl.q,wl.p)
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

	for _,item := range WaitingLine_list {
		for i := 0 ; i < item.n_serv ; i++ {
			item.t_dep[i] = t_max+1
		}
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