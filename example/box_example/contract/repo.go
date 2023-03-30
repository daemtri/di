package contract

type UserRepository interface {
	GetUserProfile(userid string) (*UserProfile, error)
}

type UserProfile struct {
	UserID   string
	Nickname string
	Email    string
	Avatar   string
}
