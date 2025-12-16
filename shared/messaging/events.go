package messaging

import (
	pbd "CarpoolSharing/shared/proto/driver"
	pb "CarpoolSharing/shared/proto/trip"
)

const (
	FindAvailableDriversQueue             = "find_available_drivers"
	DriverCmdTripRequestQueue             = "driver_cmd_trip_request"
	DriverTripResponseQueue               = "driver_trip_response"
	NotifyDriversNoDriverFoundQueue       = "notify_drivers_no_driver_found"
	NotifyDriverAssignQueue               = "notify_driver_assign"
	PaymentTripResponseQueue              = "payment_trip_response"
	NotifyPaymentEventSessionCreatedQueue = "notify_payment_event_session_created"
	NotifyPaymentSuccessQueue             = "notify_payment_success"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}

type DriverTripResponseData struct {
	Driver  *pbd.Driver `json:"driver"`
	TripID  string      `json:"tripID"`
	RiderID string      `json:"riderID"`
}

type PaymentTripResponseData struct {
	TripID   string  `json:"tripID"`
	UserID   string  `json:"userID"`
	DriverID string  `json:"driverID"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type PaymentEventSessionCreatedData struct {
	TripID    string  `json:"tripID"`
	SessionID string  `json:"sessionID"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

type PaymentStatusUpdateData struct {
	TripID   string `json:"tripID"`
	UserID   string `json:"userID"`
	DriverID string `json:"driverID"`
}
