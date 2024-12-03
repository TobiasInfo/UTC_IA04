package models

type DeadRequest struct {
	MemberID     int
	MemberType   string
	ResponseChan chan DeadResponse
}

type DeadResponse struct {
	Authorized bool
}
