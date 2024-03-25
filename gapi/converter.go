package gapi

import (
	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:         user.Username,
		FullName:         user.FullName,
		Email:            user.Email,
		PaswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreatedAt:        timestamppb.New(user.CreatedAt),
	}
}

func convertAccount(account db.Account) *pb.Account {
	return &pb.Account{
		Id:       account.ID,
		Owner:    account.Owner,
		Balance:  account.Balance,
		Carrency: account.Carrency,
	}
}

func convertAccounts(accounts []db.Account) (pbAccs []*pb.Account) {
	for _, account := range accounts {
		pbAccs = append(pbAccs, &pb.Account{
			Id:       account.ID,
			Owner:    account.Owner,
			Balance:  account.Balance,
			Carrency: account.Carrency,
		})
	}
	return
}
