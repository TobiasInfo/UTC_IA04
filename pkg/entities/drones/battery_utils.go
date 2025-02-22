package drones

import (
	"UTC_IA04/pkg/models"
	"fmt"
	"math/rand"
)

func (d *Drone) BatteryManagement() (models.Position, bool) {
	closestStation, minDistance := d.closestPOI(models.ChargingStation)
	if d.Battery <= minDistance+5 || d.DroneState == GoingToCharge || d.DroneState == FinalGoingToDock {
		step := d.nextStepToPos(closestStation)
		d.DroneState = GoingToCharge
		return step, true
	}
	return models.Position{}, false
}

func (d *Drone) tryCharging() bool {
	if d.DroneState != GoingToCharge && d.DroneState != FinalGoingToDock && !d.IsCharging {
		return false
	}

	if d.IsCharging {
		d.Battery += 5
		if d.DroneState == FinalGoingToDock {
			return true
		}
		if d.Battery >= 80+rand.Float64()*20 {
			d.IsCharging = false
			d.DroneState = NoDefinedState
			return false
		}
		return true
	}

	responseChan := make(chan models.ChargingResponse)
	d.ChargingChan <- models.ChargingRequest{
		DroneID:      d.ID,
		Position:     d.Position,
		ResponseChan: responseChan,
	}

	response := <-responseChan
	if response.Authorized {
		fmt.Printf("[DRONE %d] Starting to charge at (%.0f, %.0f)\n", d.ID, d.Position.X, d.Position.Y)
		d.IsCharging = true
		d.Battery += 5
		return true
	}
	return false
}
