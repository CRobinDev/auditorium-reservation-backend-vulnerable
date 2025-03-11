package enum

type ConferenceStatus string

const (
	ConferencePending  ConferenceStatus = "pending"
	ConferenceApproved ConferenceStatus = "approved"
	ConferenceRejected ConferenceStatus = "rejected"
)

func (s ConferenceStatus) String() string {
	return string(s)
}
