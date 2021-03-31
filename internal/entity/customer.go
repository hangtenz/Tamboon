package entity

import "time"

//Cutomer defined for transaction with each customer request
type Customer struct {
	Name            string
	AmountSubunits  int64
	CCNumber        string
	SecurityCode    string
	ExpirationMonth time.Month
	ExpirationYear  int
}

func NewCustomer(name string, amount int64, cc string, cvv string, expMonth time.Month, expYear int) Customer {
	return Customer{
		Name:            name,
		AmountSubunits:  amount,
		CCNumber:        cc,
		SecurityCode:    cvv,
		ExpirationMonth: expMonth,
		ExpirationYear:  expYear,
	}
}

//CustomerDonation defined for total donation by each customer
type CustomerDonation struct {
	Name  string
	Total int64
}

type Donations []CustomerDonation

func (d Donations) Len() int {
	return len(d)
}
func (d Donations) Less(i, j int) bool {
	return d[i].Total < d[j].Total
}
func (d Donations) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
