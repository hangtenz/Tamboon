package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"omise/go-tamboon/config"
	"omise/go-tamboon/internal/entity"
	"omise/go-tamboon/internal/helper"
	"omise/go-tamboon/internal/tamboon"
	"omise/go-tamboon/lib/cipher"
	"os"
	"sort"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		fileName       string
		file           *os.File
		reader         *cipher.Rot128Reader
		customer       entity.Customer
		ctx            = context.Background()
		err            error
		usecase        *tamboon.Usecase
		wg             sync.WaitGroup
		env            *config.Env
		delimiter      = byte('\n')
		totalDonation  = int64(0)
		sucessDonation = int64(0)
		failDonation   = int64(0)
	)

	env = config.LoadDevEnv()
	usecase = tamboon.NewUsecase(env.OmisePublicKey, env.OmiseSecretKey)
	if len(os.Args) < 2 {
		fmt.Println("Usage:./tamboon [/path/to/file]")
		fmt.Println("exiting...")
		return
	}
	fileName = os.Args[1]
	file, err = os.Open(fileName)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	customerChan := make(chan entity.Customer, env.MaxGoRoutine)
	donationByCustomer := make(map[string]int64)
	reader, err = cipher.NewRot128Reader(*bufio.NewReader(file))
	delimiter -= 128
	fmt.Println("Performing donations...")
	for {
		data, err := reader.ReadBytes(delimiter)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Error(err)
		}
		record := strings.Split(strings.TrimSpace(string(data)), ",")
		if record[0] == "Name" { //skip header
			continue
		}
		customer = entity.NewCustomer(record[0], helper.MustParseInt64(record[1]), record[2], record[3], helper.MustParseMonth(record[4]), helper.MustParseInt(record[5]))
		customerChan <- customer
		wg.Add(1)
		go perform(ctx, usecase, &wg, &customerChan, &donationByCustomer, &failDonation, &sucessDonation, &totalDonation)
	}

	wg.Wait()
	display(totalDonation, sucessDonation, failDonation, donationByCustomer)
}

func perform(ctx context.Context, usecase *tamboon.Usecase, wg *sync.WaitGroup, customerChan *chan entity.Customer, donation *map[string]int64, failDonation *int64, sucessDonation *int64, totalDonation *int64) {
	var (
		err error
	)
	defer wg.Done()
	c := <-(*customerChan)
	err = usecase.Tamboon(ctx, c)
	if err != nil {
		(*failDonation) += c.AmountSubunits
		log.Errorf("Failed when create transaction with customer %v get error %v", c.Name, err)
	} else {
		v, have := (*donation)[c.Name]
		if !have {
			(*donation)[c.Name] = c.AmountSubunits
		} else {
			(*donation)[c.Name] = v + c.AmountSubunits
		}
		(*sucessDonation) += c.AmountSubunits
	}
	(*totalDonation) += c.AmountSubunits
}

func display(totalDonation, sucessDonation, failDonation int64, donationByCustomer map[string]int64) {
	//Display top donation
	fmt.Println("done.")
	fmt.Println("total received: THB ", totalDonation)
	fmt.Println("successfully donated: THB ", sucessDonation)
	fmt.Println("faulty donation: THB ", failDonation)
	if len(donationByCustomer) == 0 {
		return
	}
	fmt.Println("average per person: THB ", int(sucessDonation)/len(donationByCustomer))

	ListDonation := entity.Donations{}
	for customerName, totalperCustomer := range donationByCustomer {
		ListDonation = append(ListDonation, entity.CustomerDonation{
			Name:  customerName,
			Total: totalperCustomer,
		})
	}
	sort.Sort(sort.Reverse(ListDonation))
	fmt.Println("top donors: ")
	for rank := 0; rank < 3; rank++ {
		if len(ListDonation) <= rank {
			break
		}
		fmt.Println(ListDonation[rank].Name)
	}
}
