package repository

type User struct {
}

func NewUser() *User {
	return &User{}
}

func (u *User) Exists() bool { 
	return false
}

func (u *User) Save(login string, password string) error {
    return nil
}