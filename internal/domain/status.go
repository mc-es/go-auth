package domain

type Status string

const (
	StatusPending Status = "pending"
	StatusActive  Status = "active"
	StatusBanned  Status = "banned"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	return s == StatusPending || s == StatusActive || s == StatusBanned
}
