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
	"time"

	log "github.com/sirupsen/logrus"
)

type CountDonation struct {
	totalDonation  int64
	sucessDonation int64
	failDonation   int64
}

var (
	rateMultiple = 1
	rateLimit    = time.Tick(time.Duration(rateMultiple) * time.Millisecond)
	customerChan = make(chan entity.Customer, 2) //Default goroutine is 2
	wg           sync.WaitGroup
)

func main() {
	var (
		fileName      string
		file          *os.File
		reader        *cipher.Rot128Reader
		customer      entity.Customer
		ctx           = context.Background()
		err           error
		usecase       *tamboon.Usecase
		env           *config.Env
		delimiter     = byte('\n')
		countDonation = CountDonation{}
	)

	//TODO: manual config key
	env = config.LoadDevEnv()
	customerChan = make(chan entity.Customer, env.MaxGoRoutine)
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
		go perform(ctx, usecase, &donationByCustomer, &countDonation)
	}

	wg.Wait()
	display(&countDonation, donationByCustomer)
}

func perform(ctx context.Context, usecase *tamboon.Usecase, donation *map[string]int64, countDonation *CountDonation) {
	var (
		err error
	)
	defer wg.Done()

	<-rateLimit
	c := <-customerChan
	err = usecase.Tamboon(ctx, c)
	if err != nil {
		//ถ้าเจอ error ให้เพิ่มเวลาใน timeTick ทีละ 100 ms
		rateMultiple += 100
		rateLimit = time.Tick(time.Duration(rateMultiple) * time.Millisecond)
		(countDonation.failDonation) += c.AmountSubunits
		log.Errorf("Failed when create transaction with customer %v get error %v", c.Name, err)
	} else {
		v, have := (*donation)[c.Name]
		if !have {
			(*donation)[c.Name] = c.AmountSubunits
		} else {
			(*donation)[c.Name] = v + c.AmountSubunits
		}
		(countDonation.sucessDonation) += c.AmountSubunits
	}
	(countDonation.totalDonation) += c.AmountSubunits
}

func display(countDonation *CountDonation, donationByCustomer map[string]int64) {
	//TODO: fix show money
	//Display top donation
	fmt.Println("done.")
	fmt.Println("total received: THB ", countDonation.totalDonation)
	fmt.Println("successfully donated: THB ", countDonation.sucessDonation)
	fmt.Println("faulty donation: THB ", countDonation.failDonation)
	if len(donationByCustomer) == 0 {
		return
	}
	fmt.Println("average per person: THB ", int(countDonation.sucessDonation)/len(donationByCustomer))

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
