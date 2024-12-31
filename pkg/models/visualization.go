package models

type DensityGrid struct {
	Grid     [][]float64
	Width    int
	Height   int
	CellSize int
}

type DroneNetwork struct {
	DronePositions    []Position
	DroneConnections  []Position
	RescueConnections []Position
}
