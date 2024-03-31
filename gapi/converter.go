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

func convertTransfer(transfer db.Transfer) *pb.Transfer {
	return &pb.Transfer{
		Id:            transfer.ID,
		FromAccountId: transfer.FromAccountID.Int64,
		ToAccountId:   transfer.ToAccountID.Int64,
		Amount:        transfer.Amount,
		CreatedAt:     timestamppb.New(transfer.CreatedAt),
	}
}

func convertEntry(entry db.Entry) *pb.Entry {
	return &pb.Entry{
		Id:        entry.ID,
		AccountId: entry.AccountID.Int64,
		Amount:    entry.Amount,
		CreatedAt: timestamppb.New(entry.CreatedAt),
	}
}

func convertTransferTx(txResp db.TransferTxResult) *pb.CreateTransferTxResponse {
	return &pb.CreateTransferTxResponse{
		Transfer:    convertTransfer(txResp.Transfer),
		FromAccount: convertAccount(txResp.FromAccount),
		ToAccount:   convertAccount(txResp.ToAccount),
		FromEntry:   convertEntry(txResp.FromEntry),
		ToEntry:     convertEntry(txResp.ToEntry),
	}
}
