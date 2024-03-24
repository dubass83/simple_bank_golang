package val

import (
	"fmt"
	"net/mail"
	"regexp"

	"github.com/dubass83/simplebank/util"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	isValidFullName = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

func validateString(val string, min, max int) error {
	n := len(val)
	if n < min || n > max {
		return fmt.Errorf("string lenght is %d less then min: %d or more then max: %d", n, min, max)
	}
	return nil
}

func ValidateUsername(username string) error {
	if err := validateString(username, 4, 20); err != nil {
		return err
	}
	if !isValidUsername(username) {
		return fmt.Errorf("must contain only low case letters number and _")
	}
	return nil
}

func ValidateFullname(fullName string) error {
	if err := validateString(fullName, 4, 100); err != nil {
		return err
	}
	if !isValidFullName(fullName) {
		return fmt.Errorf("must contain only letters and spaces")
	}
	return nil
}

func ValidatePassword(password string) error {
	return validateString(password, 8, 21)
}

func ValidateEmail(email string) error {
	if err := validateString(email, 4, 200); err != nil {
		return err
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("is not a valid email")
	}
	return nil
}

func ValidateAccountCurrency(currrency string) error {
	if !util.IfSupportedCurrency(currrency) {
		return fmt.Errorf("not supported currency: %s", currrency)
	}
	return nil
}

func ValidateAccountId(id int64) error {
	if id < 1 {
		return fmt.Errorf("id must be a positive value gt then 0")
	}
	return nil
}
