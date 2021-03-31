package gateway

import (
	"context"
	"omise/go-tamboon/internal/entity"

	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
	log "github.com/sirupsen/logrus"
)

type Store struct {
	client *omise.Client
}

func NewStore(client *omise.Client) *Store {
	return &Store{
		client: client,
	}
}

func (s *Store) Charge(ctx context.Context, customer entity.Customer) error {
	var (
		card          = &omise.Card{}
		charge        = &omise.Charge{}
		tokenCustomer = &operations.CreateToken{
			Name:            customer.Name,
			Number:          customer.CCNumber,
			ExpirationMonth: customer.ExpirationMonth,
			ExpirationYear:  customer.ExpirationYear,
			SecurityCode:    customer.SecurityCode,
		}
		transaction = &operations.CreateCharge{
			Amount:   customer.AmountSubunits,
			Currency: "THB",
		}
		err error
	)

	err = s.client.Do(card, tokenCustomer)
	if err != nil {
		return err
	}
	transaction.Card = card.ID
	err = s.client.Do(charge, transaction)
	if err != nil {
		return err
	}
	log.Infof("transaction for customer name %v have status %v with transaction id = %v", customer.Name, charge.Status, charge.Transaction)
	return nil
}
