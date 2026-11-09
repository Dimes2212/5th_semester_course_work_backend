package user

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrFullNameRequired = errors.New("full name required")
var ErrFullNameTooLong = errors.New("full name too long")
