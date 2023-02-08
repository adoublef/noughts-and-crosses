package reg

import "github.com/google/uuid"

type Profile struct {
	ID       uuid.UUID
	Email    string
	Username string
	Bio      string
	PhotoURL string
}
