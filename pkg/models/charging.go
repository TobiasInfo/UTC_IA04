package models

type ChargingRequest struct {
    DroneID      int
    Position     Position
    ResponseChan chan ChargingResponse
}

type ChargingResponse struct {
    Authorized bool
    Reason     string
}