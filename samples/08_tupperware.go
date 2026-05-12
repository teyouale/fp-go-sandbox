// Chapter 08: Tupperware — Option, Either, and IO
package main

import (
	"fmt"
	"time"

	E "github.com/IBM/fp-go/either"
	"github.com/IBM/fp-go/errors"
	F "github.com/IBM/fp-go/function"
	I "github.com/IBM/fp-go/identity"
	N "github.com/IBM/fp-go/number"
	O "github.com/IBM/fp-go/option"
	"github.com/IBM/fp-go/ord"
	S "github.com/IBM/fp-go/string"
)

type Account struct {
	Balance float32
}

func MakeAccount(b float32) Account { return Account{Balance: b} }
func getBalance(a Account) float32  { return a.Balance }

var (
	ordFloat32       = ord.FromStrictCompare[float32]()
	UpdateLedger     = F.Identity[Account]
	RemainingBalance = F.Flow2(getBalance, S.Format[float32]("Your balance is $%0.2f"))
	FinishTransaction = F.Flow2(UpdateLedger, RemainingBalance)
)

func Withdraw(amount float32) func(account Account) O.Option[Account] {
	return F.Flow3(
		getBalance,
		O.FromPredicate(ord.Geq(ordFloat32)(amount)),
		O.Map(F.Flow2(N.Add(-amount), MakeAccount)),
	)
}

type User struct {
	BirthDate string
}

func getBirthDate(u User) string { return u.BirthDate }
func MakeUser(d string) User     { return User{BirthDate: d} }

var parseDate = F.Bind1of2(E.Eitherize2(time.Parse))(time.DateOnly)

func GetAge(now time.Time) func(User) E.Either[error, float64] {
	return F.Flow3(
		getBirthDate,
		parseDate,
		E.Map[error](F.Flow3(now.Sub, time.Duration.Hours, N.Mul(1/24.0))),
	)
}

func main() {
	fmt.Println("=== Chapter 08: Tupperware (Functors) ===")
	fmt.Println()

	// Withdraw example with Option
	getTwenty := F.Flow2(
		Withdraw(20),
		O.Fold(F.Constant("You're broke!"), FinishTransaction),
	)

	fmt.Println("Withdraw $20 from $200:", getTwenty(MakeAccount(200)))
	fmt.Println("Withdraw $20 from $10: ", getTwenty(MakeAccount(10)))
	fmt.Println()

	// GetAge with Either
	now, _ := time.Parse(time.DateOnly, "2023-09-01")

	fmt.Println("GetAge(valid date):  ", GetAge(now)(MakeUser("2005-12-12")))
	fmt.Println("GetAge(invalid date):", GetAge(now)(MakeUser("July 4, 2001")))
	fmt.Println()

	// Zoltar: composing Either transformations
	Concat := F.Curry2(S.Monoid.Concat)
	fortune := F.Flow3(
		N.Add(365.0),
		S.Format[float64]("%0.0f"),
		Concat("If you survive, you will be "),
	)

	zoltar := F.Flow3(
		GetAge(now),
		E.Map[error](fortune),
		E.GetOrElse(errors.ToString),
	)

	fmt.Println("Zoltar says:", zoltar(MakeUser("2005-12-12")))
	fmt.Println()

	// Identity functor
	incrF := I.Map(N.Add(1))
	fmt.Println("Identity Map increment(2):", incrF(I.Of(2)))
}
