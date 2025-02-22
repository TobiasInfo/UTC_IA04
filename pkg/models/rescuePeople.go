package models

type RescuePeopleRequest struct {
	PersonID      int
	RescuerID     int
	RescuePointID int
	DroneSenderID int
	ResponseChan  chan RescuePeopleResponse
}

type RescuePeopleResponse struct {
	Authorized bool
	Reason     string
}
