package model

import "context"

// User
type User struct {
	Base
	PreferedUserName string `json:"prefered_user_name"`
	GivenName        string `json:"given_name"`
	FamilyName       string `json:"family_name"`
	Email            string `json:"email"`
}

func (u *User) ResourceName() string {
	return "user"
}

// Validate checks structure consistency
func (u *User) Validate(ctx context.Context) error {
	return nil
}

func (u *User) Prepare(ctx context.Context) error {
	err := u.BasePrepare()
	if err != nil {
		return err
	}
	return nil
}
