package models

type MedicalDeliveryRequest struct {
	PersonID     int
	DroneID      int
	ResponseChan chan MedicalDeliveryResponse
}

type MedicalDeliveryResponse struct {
	Authorized bool
	Reason     string
}