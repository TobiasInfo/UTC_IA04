package main

import (
	"UTC_IA04/pkg/simulation"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type SimulationConfig struct {
	NumDrones int
	NumPeople int
	Protocol  int
	MapName   string
}

type AggregatedMetrics struct {
	TotalPeople     float64
	InDistress      float64
	CasesTreated    float64
	CasesDead       float64
	AverageBattery  float64
	AverageCoverage float64
	Runtime         time.Duration
	TotalTicks      int
	RescueStats     simulation.SimulationRescueStats
}

func main() {
	// Get the directory where the program is running
	execDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("Error getting executable directory: %v\n", err)
		return
	}

	// Create results directory inside the run_simulations folder
	resultsDir := filepath.Join(execDir, "results")
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		fmt.Printf("Error creating results directory: %v\n", err)
		return
	}

	fmt.Println("Starting simulation runs...")
	fmt.Printf("Results will be stored in: %s\n", resultsDir)

	droneConfigs := []int{2, 5, 10}
	peopleConfigs := []int{200, 500, 1000}
	protocolConfigs := []int{1, 2, 3, 4}
	mapConfigs := []string{"festival_layout_1", "festival_layout_2", "festival_layout_3"}

	for _, drones := range droneConfigs {
		for _, people := range peopleConfigs {
			for _, protocol := range protocolConfigs {
				for _, mapName := range mapConfigs {
					config := SimulationConfig{
						NumDrones: drones,
						NumPeople: people,
						Protocol:  protocol,
						MapName:   mapName,
					}

					dirName := fmt.Sprintf("%dd_%dp_p%d_%s", drones, people, protocol, mapName)
					configDir := filepath.Join(resultsDir, dirName)
					fmt.Printf("\n===== Starting configuration: %s =====\n", dirName)
					runSimulationSeries(config, configDir)
				}
			}
		}
	}
	fmt.Println("\nAll simulation runs completed!")
}

func runSimulationSeries(config SimulationConfig, resultDir string) {
	if err := os.MkdirAll(resultDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", resultDir, err)
		return
	}

	var metrics []AggregatedMetrics

	// Run 5 iterations sequentially
	for i := 0; i < 5; i++ {
		fmt.Printf("Starting run %d/5\n", i+1)
		metric := runSingleSimulation(config, i)
		metrics = append(metrics, metric)
		fmt.Printf("Completed run %d/5 (took %d ticks)\n", i+1, metric.TotalTicks)

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Averaging metrics and generating final outputs...\n")
	avgMetrics := averageMetrics(metrics)
	exportResults(avgMetrics, resultDir)
	plotAggregatedResults(avgMetrics, resultDir)
	fmt.Printf("Results exported to: %s\n", resultDir)
}

func runSingleSimulation(config SimulationConfig, runNum int) AggregatedMetrics {
	startTime := time.Now()

	// Create new simulation with defaults
	sim := simulation.NewSimulation(0, 0, 0)

	// Update configuration in correct order
	fmt.Printf("Loading Map %s\n", config.MapName)
	sim.UpdateMap(config.MapName)

	fmt.Printf("Updating configuration - Drones: %d, People: %d, Protocol: %d\n",
		config.NumDrones, config.NumPeople, config.Protocol)
	sim.UpdateDroneSize(config.NumDrones)
	sim.UpdateCrowdSize(config.NumPeople)

	fmt.Printf("Setting protocol to %d\n", config.Protocol)
	sim.UpdateDroneProtocole(config.Protocol)

	fmt.Printf("Initializing drone protocols\n")
	sim.InitDronesProtocols()

	tick := 0
	for {
		sim.Update()
		tick++

		if tick%100 == 0 {
			fmt.Printf("Run %d: Tick %d\n", runNum+1, tick)
		}

		allPeopleOut := true
		for _, person := range sim.Persons {
			if person.StillInSim {
				allPeopleOut = false
				break
			}
		}

		if allPeopleOut {
			allDronesCharging := true
			for _, drone := range sim.Drones {
				if !drone.IsCharging {
					allDronesCharging = false
					break
				}
			}

			if allDronesCharging {
				fmt.Printf("Run %d: Natural completion at tick %d\n", runNum+1, tick)
				break
			}
		}
	}

	stats := sim.GetStatistics()
	return AggregatedMetrics{
		TotalPeople:     float64(stats.TotalPeople),
		InDistress:      float64(stats.InDistress),
		CasesTreated:    float64(stats.CasesTreated),
		CasesDead:       float64(stats.CasesDead),
		AverageBattery:  stats.AverageBattery,
		AverageCoverage: stats.AverageCoverage,
		Runtime:         time.Since(startTime),
		TotalTicks:      tick,
		RescueStats:     sim.SimulationRescueStats,
	}
}

func plotAggregatedResults(metrics AggregatedMetrics, dirPath string) {
	// Plot people data (distress and rescued)
	p1 := plot.New()
	p1.Title.Text = "Festival Rescue Statistics - People"
	p1.X.Label.Text = "Ticks"
	p1.Y.Label.Text = "Number of People"

	// Convert map data to plottable points
	ticks := make([]float64, 0)
	distressData := make([]float64, 0)
	rescuedData := make([]float64, 0)

	// Get all ticks in order
	for tick := range metrics.RescueStats.PersonsInDistress {
		ticks = append(ticks, float64(tick))
	}
	sort.Float64s(ticks)

	// Create data points
	for _, tick := range ticks {
		t := int(tick)
		distressData = append(distressData, float64(metrics.RescueStats.PersonsInDistress[t]))
		rescuedData = append(rescuedData, float64(metrics.RescueStats.PersonsRescued[t]))
	}

	// Create and style distress line
	distressPoints := make(plotter.XYs, len(ticks))
	for i := range ticks {
		distressPoints[i].X = ticks[i]
		distressPoints[i].Y = distressData[i]
	}
	distressLine, err := plotter.NewLine(distressPoints)
	if err == nil {
		distressLine.Color = color.RGBA{R: 255, A: 255} // Red
		distressLine.Width = vg.Points(1)
		p1.Add(distressLine)
		p1.Legend.Add("In Distress", distressLine)
	}

	// Create and style rescued line
	rescuedPoints := make(plotter.XYs, len(ticks))
	for i := range ticks {
		rescuedPoints[i].X = ticks[i]
		rescuedPoints[i].Y = rescuedData[i]
	}
	rescuedLine, err := plotter.NewLine(rescuedPoints)
	if err == nil {
		rescuedLine.Color = color.RGBA{G: 255, A: 255} // Green
		rescuedLine.Width = vg.Points(1)
		p1.Add(rescuedLine)
		p1.Legend.Add("Rescued", rescuedLine)
	}

	// Save people plot
	peopleFile := filepath.Join(dirPath, "rescue_stats_people.png")
	if err := p1.Save(8*vg.Inch, 4*vg.Inch, peopleFile); err != nil {
		fmt.Printf("Error saving people plot: %v\n", err)
	}

	// Plot average rescue time
	p2 := plot.New()
	p2.Title.Text = "Festival Rescue Statistics - Average Rescue Time"
	p2.X.Label.Text = "Ticks"
	p2.Y.Label.Text = "Time (ticks)"

	// Calculate average rescue times
	avgTimeData := make([]float64, 0)
	for _, tick := range ticks {
		t := int(tick)
		times := metrics.RescueStats.AvgRescueTime[t]
		if len(times) > 0 {
			sum := 0
			for _, time := range times {
				sum += time
			}
			avgTimeData = append(avgTimeData, float64(sum)/float64(len(times)))
		} else {
			avgTimeData = append(avgTimeData, 0)
		}
	}

	// Create and style average time line
	avgTimePoints := make(plotter.XYs, len(ticks))
	for i := range ticks {
		avgTimePoints[i].X = ticks[i]
		avgTimePoints[i].Y = avgTimeData[i]
	}
	avgTimeLine, err := plotter.NewLine(avgTimePoints)
	if err == nil {
		avgTimeLine.Color = color.RGBA{B: 255, A: 255} // Blue
		avgTimeLine.Width = vg.Points(1)
		p2.Add(avgTimeLine)
		p2.Legend.Add("Avg Rescue Time", avgTimeLine)
	}

	// Save time plot
	timeFile := filepath.Join(dirPath, "rescue_stats_time.png")
	if err := p2.Save(8*vg.Inch, 4*vg.Inch, timeFile); err != nil {
		fmt.Printf("Error saving time plot: %v\n", err)
	}
}

func averageMetrics(metrics []AggregatedMetrics) AggregatedMetrics {
	var avg AggregatedMetrics
	count := float64(len(metrics))

	// Initialize maps in avg.RescueStats
	avg.RescueStats.PersonsInDistress = make(map[int]int)
	avg.RescueStats.PersonsRescued = make(map[int]int)
	avg.RescueStats.AvgRescueTime = make(map[int][]int)

	for _, m := range metrics {
		avg.TotalPeople += m.TotalPeople
		avg.InDistress += m.InDistress
		avg.CasesTreated += m.CasesTreated
		avg.CasesDead += m.CasesDead
		avg.AverageBattery += m.AverageBattery
		avg.AverageCoverage += m.AverageCoverage
		avg.Runtime += m.Runtime
		avg.TotalTicks += m.TotalTicks

		// Merge rescue stats
		for tick, value := range m.RescueStats.PersonsInDistress {
			avg.RescueStats.PersonsInDistress[tick] += value
		}
		for tick, value := range m.RescueStats.PersonsRescued {
			avg.RescueStats.PersonsRescued[tick] += value
		}
		for tick, times := range m.RescueStats.AvgRescueTime {
			avg.RescueStats.AvgRescueTime[tick] = append(avg.RescueStats.AvgRescueTime[tick], times...)
		}
	}

	// Calculate averages
	avg.TotalPeople /= count
	avg.InDistress /= count
	avg.CasesTreated /= count
	avg.CasesDead /= count
	avg.AverageBattery /= count
	avg.AverageCoverage /= count
	avg.Runtime /= time.Duration(count)
	avg.TotalTicks = int(float64(avg.TotalTicks) / count)

	// Average rescue stats
	for tick := range avg.RescueStats.PersonsInDistress {
		avg.RescueStats.PersonsInDistress[tick] = int(float64(avg.RescueStats.PersonsInDistress[tick]) / count)
	}
	for tick := range avg.RescueStats.PersonsRescued {
		avg.RescueStats.PersonsRescued[tick] = int(float64(avg.RescueStats.PersonsRescued[tick]) / count)
	}

	return avg
}

func exportResults(metrics AggregatedMetrics, dirPath string) {
	content := fmt.Sprintf(`Simulation Results (Averaged over 5 runs)
=====================================
Total People: %.2f
People in Distress: %.2f
Cases Treated: %.2f
Cases Dead: %.2f
Average Battery: %.2f%%
Average Coverage: %.2f%%
Average Runtime: %v
Total Ticks: %d

Performance Metrics:
- Treatment Success Rate: %.2f%%
- Mortality Rate: %.2f%%
- Average Response Time: %v
`,
		metrics.TotalPeople,
		metrics.InDistress,
		metrics.CasesTreated,
		metrics.CasesDead,
		metrics.AverageBattery,
		metrics.AverageCoverage,
		metrics.Runtime,
		metrics.TotalTicks,
		(metrics.CasesTreated/metrics.InDistress)*100,
		(metrics.CasesDead/metrics.InDistress)*100,
		metrics.Runtime/time.Duration(metrics.CasesTreated),
	)

	metricsPath := filepath.Join(dirPath, "metrics.txt")
	if err := os.WriteFile(metricsPath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing metrics file: %v\n", err)
	}
}
