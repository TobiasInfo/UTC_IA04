package models

type MovementRequest struct {
	MemberID     int
	MemberType   string
	NewPosition  Position
	ResponseChan chan MovementResponse
}

type MovementResponse struct {
	Authorized bool
}
