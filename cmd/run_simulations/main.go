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
	// Create results directory in the current project directory
	resultsDir := filepath.Join(".", "results")
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		fmt.Printf("Error creating results directory: %v\n", err)
		return
	}

	fmt.Println("Starting simulation runs...")
	fmt.Printf("Results will be stored in: %s\n", resultsDir)

	// Configuration parameters
	droneConfigs := []int{2, 5, 10}
	peopleConfigs := []int{200, 500, 1000}
	protocolConfigs := []int{1, 2, 3, 4}
	mapConfigs := []string{"festival_layout_1", "festival_layout_2", "festival_layout_3"}

	// Run simulations for each configuration
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

					// Create directory for this configuration
					if err := os.MkdirAll(configDir, 0755); err != nil {
						fmt.Printf("Error creating directory for configuration: %v\n", err)
						continue
					}

					runSimulationSeries(config, configDir)
				}
			}
		}
	}
	fmt.Println("\nAll simulation runs completed!")
}

func runSimulationSeries(config SimulationConfig, resultDir string) {
	var metrics []AggregatedMetrics

	// Run 5 iterations
	for i := 0; i < 5; i++ {
		fmt.Printf("Starting run %d/5\n", i+1)
		metric := runSingleSimulation(config, i)
		metrics = append(metrics, metric)
		fmt.Printf("Completed run %d/5 (took %d ticks)\n", i+1, metric.TotalTicks)

		// Save individual run results
		saveRunResults(metric, resultDir, i+1)

	}

	// Process and save aggregated results
	fmt.Printf("Averaging metrics and generating final outputs...\n")
	avgMetrics := averageMetrics(metrics)
	exportResults(avgMetrics, resultDir)
	plotAggregatedResults(avgMetrics, resultDir)
	fmt.Printf("Results exported to: %s\n", resultDir)
}

func runSingleSimulation(config SimulationConfig, runNum int) AggregatedMetrics {
	startTime := time.Now()

	// Initialize simulation with defaults
	sim := simulation.NewSimulation(0, 0, 0)

	// Configure simulation
	sim.UpdateMap(config.MapName)
	sim.UpdateDroneSize(config.NumDrones)
	sim.UpdateCrowdSize(config.NumPeople)
	sim.UpdateDroneProtocole(config.Protocol)
	sim.InitDronesProtocols()

	tick := 0
	for {
		sim.Update()
		tick++

		// Progress update every 100 ticks
		if tick%100 == 0 {
			fmt.Printf("Run %d: Tick %d\n", runNum+1, tick)
		}

		// Check if simulation is complete
		if isSimulationComplete(sim) {
			fmt.Printf("Run %d: Completed at tick %d\n", runNum+1, tick)
			break
		}
	}

	// Collect final statistics
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

func isSimulationComplete(sim *simulation.Simulation) bool {
	allPeopleOut := true
	for _, person := range sim.Persons {
		if person.StillInSim {
			allPeopleOut = false
			break
		}
	}

	if !allPeopleOut {
		return false
	}

	allDronesCharging := true
	for _, drone := range sim.Drones {
		if !drone.IsCharging {
			allDronesCharging = false
			break
		}
	}

	return allDronesCharging
}

func saveRunResults(metrics AggregatedMetrics, dirPath string, runNum int) {
	content := fmt.Sprintf(`Run %d Results
================
Total People: %.2f
People in Distress: %.2f
Cases Treated: %.2f
Cases Dead: %.2f
Average Battery: %.2f%%
Average Coverage: %.2f%%
Runtime: %v
Total Ticks: %d
`,
		runNum,
		metrics.TotalPeople,
		metrics.InDistress,
		metrics.CasesTreated,
		metrics.CasesDead,
		metrics.AverageBattery,
		metrics.AverageCoverage,
		metrics.Runtime,
		metrics.TotalTicks,
	)

	filename := fmt.Sprintf("run_%d_metrics.txt", runNum)
	filepath := filepath.Join(dirPath, filename)
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		fmt.Printf("Error saving run %d results: %v\n", runNum, err)
	}
}

func plotAggregatedResults(metrics AggregatedMetrics, dirPath string) {
	// Create people statistics plot
	p1 := plot.New()
	p1.Title.Text = "Festival Rescue Statistics - People"
	p1.X.Label.Text = "Ticks"
	p1.Y.Label.Text = "Number of People"

	// Collect and sort ticks
	ticks := make([]float64, 0)
	for tick := range metrics.RescueStats.PersonsInDistress {
		ticks = append(ticks, float64(tick))
	}
	sort.Float64s(ticks)

	// Prepare data series
	distressData := make(plotter.XYs, len(ticks))
	rescuedData := make(plotter.XYs, len(ticks))
	for i, tick := range ticks {
		t := int(tick)
		distressData[i].X = tick
		distressData[i].Y = float64(metrics.RescueStats.PersonsInDistress[t])
		rescuedData[i].X = tick
		rescuedData[i].Y = float64(metrics.RescueStats.PersonsRescued[t])
	}

	// Add distress line
	if distressLine, err := plotter.NewLine(distressData); err == nil {
		distressLine.Color = color.RGBA{R: 255, A: 255}
		distressLine.Width = vg.Points(1)
		p1.Add(distressLine)
		p1.Legend.Add("In Distress", distressLine)
	}

	// Add rescued line
	if rescuedLine, err := plotter.NewLine(rescuedData); err == nil {
		rescuedLine.Color = color.RGBA{G: 255, A: 255}
		rescuedLine.Width = vg.Points(1)
		p1.Add(rescuedLine)
		p1.Legend.Add("Rescued", rescuedLine)
	}

	// Save people plot
	if err := p1.Save(8*vg.Inch, 4*vg.Inch, filepath.Join(dirPath, "rescue_stats_people.png")); err != nil {
		fmt.Printf("Error saving people plot: %v\n", err)
	}

	// Create average rescue time plot
	p2 := plot.New()
	p2.Title.Text = "Festival Rescue Statistics - Average Rescue Time"
	p2.X.Label.Text = "Ticks"
	p2.Y.Label.Text = "Time (ticks)"

	// Calculate and plot average rescue times
	avgTimeData := make(plotter.XYs, len(ticks))
	for i, tick := range ticks {
		t := int(tick)
		times := metrics.RescueStats.AvgRescueTime[t]
		avgTimeData[i].X = tick
		if len(times) > 0 {
			sum := 0
			for _, time := range times {
				sum += time
			}
			avgTimeData[i].Y = float64(sum) / float64(len(times))
		}
	}

	if avgTimeLine, err := plotter.NewLine(avgTimeData); err == nil {
		avgTimeLine.Color = color.RGBA{B: 255, A: 255}
		avgTimeLine.Width = vg.Points(1)
		p2.Add(avgTimeLine)
		p2.Legend.Add("Avg Rescue Time", avgTimeLine)
	}

	// Save time plot
	if err := p2.Save(8*vg.Inch, 4*vg.Inch, filepath.Join(dirPath, "rescue_stats_time.png")); err != nil {
		fmt.Printf("Error saving time plot: %v\n", err)
	}
}

func averageMetrics(metrics []AggregatedMetrics) AggregatedMetrics {
	if len(metrics) == 0 {
		return AggregatedMetrics{}
	}

	var avg AggregatedMetrics
	count := float64(len(metrics))

	// Initialize rescue stats maps
	avg.RescueStats.PersonsInDistress = make(map[int]int)
	avg.RescueStats.PersonsRescued = make(map[int]int)
	avg.RescueStats.AvgRescueTime = make(map[int][]int)

	// Sum all metrics
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

	if err := os.WriteFile(filepath.Join(dirPath, "metrics.txt"), []byte(content), 0644); err != nil {
		fmt.Printf("Error writing metrics file: %v\n", err)
	}
}
