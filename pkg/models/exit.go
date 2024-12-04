package models

type ExitRequest struct {
	MemberID     int
	MemberType   string
	ResponseChan chan ExitResponse
}

type ExitResponse struct {
	Authorized bool
	Reason     string
}
