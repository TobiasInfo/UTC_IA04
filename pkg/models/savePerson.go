package models

type SavePersonRequest struct {
	PersonID     int
	DroneID      int
	ResponseChan chan SavePersonResponse
}

type SavePersonResponse struct {
	Authorized bool
	Reason     string
}