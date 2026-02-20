package domain

type Status string

const (
	StatusActivated Status = "activated"
	StatusBanned    Status = "banned"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	return s == StatusActivated || s == StatusBanned
}
