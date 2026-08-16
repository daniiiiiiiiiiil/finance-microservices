package domain

import (
	errors_core "backend/internal/core/errors"
	"fmt"
	"regexp"
)

type User struct {
	ID           int
	Version      int
	FullName     string
	Email        string
	PasswordHash string
	PhoneNumber  *string
	IsAdmin      bool
}

func NewUser(id int, version int, fullName string, email string, passwordHash string, phoneNumber *string, isAdmin bool) User {
	return User{
		ID:           id,
		Version:      version,
		FullName:     fullName,
		Email:        email,
		PasswordHash: passwordHash,
		PhoneNumber:  phoneNumber,
		IsAdmin:      isAdmin,
	}
}

func NewUserUninitialized(fullName, email, passwordHash string, phoneNumber *string, isAdmin bool) User {
	return NewUser(UninitializedID, UninitializedVersion, fullName, email, passwordHash, phoneNumber, isAdmin)
}

func (u *User) Validate() error {
	fullNameLength := len([]rune(u.FullName))
	if fullNameLength < 3 || fullNameLength > 100 {
		return fmt.Errorf("User full name must be between 3 and 100 characters:%d:%w", fullNameLength, errors_core.ErrInvalidArgument)
	}

	if u.Email == "" {
		return fmt.Errorf("email is required: %w", errors_core.ErrInvalidArgument)
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("invalid email format: %w", errors_core.ErrInvalidArgument)
	}

	if u.PasswordHash == "" {
		return fmt.Errorf("password hash is required: %w", errors_core.ErrInvalidArgument)
	}

	if u.PhoneNumber != nil {
		phoneNumberLength := len([]rune(*u.PhoneNumber))
		if phoneNumberLength < 10 || phoneNumberLength > 15 {
			return fmt.Errorf("invalid phone number:%d:%w", phoneNumberLength, errors_core.ErrInvalidArgument)
		}
		re := regexp.MustCompile(`^\+[0-9]+$`)
		if !re.MatchString(*u.PhoneNumber) {
			return fmt.Errorf("invalid phone number format:%w", errors_core.ErrInvalidArgument)
		}
	}

	return nil
}

type UserPatch struct {
	FullName    Nullable[string]
	PhoneNumber Nullable[string]
}

func NewUserPatch(
	fullName Nullable[string],
	phoneNumber Nullable[string]) UserPatch {
	return UserPatch{
		FullName:    fullName,
		PhoneNumber: phoneNumber,
	}
}

func (p *UserPatch) Validate() error {
	if p.FullName.Set && p.FullName.Value == nil {
		return fmt.Errorf("invalid full name to NULL:%w", errors_core.ErrInvalidArgument)
	}
	return nil
}

func (u *User) ApplyPatch(patch UserPatch) error {
	if err := patch.Validate(); err != nil {
		return fmt.Errorf("invalid patch: %w", err)
	}

	tmp := *u

	if patch.FullName.Set {
		if patch.FullName.Value == nil {
			return fmt.Errorf("full name cannot be null: %w", errors_core.ErrInvalidArgument)
		}
		tmp.FullName = *patch.FullName.Value
	}

	if patch.PhoneNumber.Set {
		tmp.PhoneNumber = patch.PhoneNumber.Value
	}

	if err := tmp.ValidatePatch(); err != nil {
		return fmt.Errorf("invalid user after patch: %w", err)
	}

	*u = tmp
	return nil
}

func (u *User) ValidatePatch() error {
	fullNameLength := len([]rune(u.FullName))
	if fullNameLength < 3 || fullNameLength > 100 {
		return fmt.Errorf("User full name must be between 3 and 100 characters:%d:%w", fullNameLength, errors_core.ErrInvalidArgument)
	}

	if u.Email == "" {
		return fmt.Errorf("email is required: %w", errors_core.ErrInvalidArgument)
	}

	if u.PhoneNumber != nil {
		phoneNumberLength := len([]rune(*u.PhoneNumber))
		if phoneNumberLength < 10 || phoneNumberLength > 15 {
			return fmt.Errorf("invalid phone number:%d:%w", phoneNumberLength, errors_core.ErrInvalidArgument)
		}
		re := regexp.MustCompile(`^\+[0-9]+$`)
		if !re.MatchString(*u.PhoneNumber) {
			return fmt.Errorf("invalid phone number format:%w", errors_core.ErrInvalidArgument)
		}
	}

	return nil
}

func (u *User) ApplyPatchRole(patch bool) error {
	tmp := *u
	if patch {
		tmp.IsAdmin = patch
	}
	if err := tmp.Validate(); err != nil {
		return fmt.Errorf("invalid user after patch: %w", err)
	}
	*u = tmp
	return nil
}
