package val

import (
	"fmt"
	"net/mail"
	"regexp"

	db "github.com/dubass83/simplebank/db/sqlc"
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

func ValidatePageNumber(param int32) error {
	if param < 1 {
		return fmt.Errorf("page number must be a positive value gt then 0")
	}
	return nil
}

func ValidatePageSize(param int32) error {
	if param > 10 || param < 5 {
		return fmt.Errorf("page size must be in interval between 5 and 10")
	}
	return nil
}

func ValidateMoneyAmmount(amount int64) error {
	if amount < 0 {
		return fmt.Errorf("cannot transfer negative amount of money")
	}
	return nil
}

func ValidateTxCarrency(fromAccount, toAccount db.Account) error {
	if fromAccount.Carrency != toAccount.Carrency {
		return fmt.Errorf(
			"user can only transfer money with the same carrency. from Account carrency: %s - to Account carrency: %s",
			fromAccount.Carrency,
			toAccount.Carrency,
		)
	}
	return nil
}

func ValidateVerifyEmailID(id int64) error {
	if id <= 0 {
		return fmt.Errorf("id should be greater then 0")
	}
	return nil
}

func ValidateVerifyEmailSecretCode(secretCode string) error {
	return validateString(secretCode, 32, 128)
}
