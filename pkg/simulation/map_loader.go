package simulation

import (
	"UTC_IA04/pkg/entities/obstacles"
	"UTC_IA04/pkg/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadFestivalConfig loads a festival configuration from a JSON file
func LoadFestivalConfig(configPath string) (*models.FestivalConfig, error) {
	// Get absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path: %v", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config models.FestivalConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

// ApplyFestivalConfig applies a festival configuration to an existing map
func (m *Map) ApplyFestivalConfig(config *models.FestivalConfig) error {
	// Clear existing obstacles
	m.Obstacles = []*obstacles.Obstacle{}

	// Create obstacles for each POI
	for i, poi := range config.POILocations {
		// Validate position
		if poi.Position.X < 0 || poi.Position.X >= float64(m.Width) ||
			poi.Position.Y < 0 || poi.Position.Y >= float64(m.Height) {
			return fmt.Errorf("invalid POI position: %v", poi.Position)
		}

		obstacle := obstacles.NewObstacle(
			i,
			poi.Position,
			poi.Type,
			poi.Capacity,
		)
		m.AddObstacle(&obstacle)

		fmt.Printf("Added POI: Type=%d, Position=(%f,%f), Capacity=%d\n",
			poi.Type, poi.Position.X, poi.Position.Y, poi.Capacity)
	}

	return nil
}
