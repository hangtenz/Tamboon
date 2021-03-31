package tamboon

import (
	"context"
	"log"
	"omise/go-tamboon/internal/entity"
	"omise/go-tamboon/internal/repository/gateway"

	"github.com/omise/omise-go"
)

type Usecase struct {
	gateWayStore *gateway.Store
}

func NewUsecase(publicKey, secretKey string) *Usecase {
	client, err := omise.NewClient(publicKey, secretKey)
	if err != nil {
		log.Fatal(err)
	}
	return &Usecase{
		gateWayStore: gateway.NewStore(client),
	}
}

func (u *Usecase) Tamboon(ctx context.Context, customer entity.Customer) error {
	var (
		err error
	)
	err = u.gateWayStore.Charge(ctx, customer)
	if err != nil {
		return err
	}
	return nil
}
