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

	"github.com/Rhymond/go-money"
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
		record        []string
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
		record = strings.Split(strings.TrimSpace(string(data)), ",")
		if record[0] == "Name" { //skip header
			continue
		}
		customer = entity.NewCustomer(record[0], helper.MustParseInt64(record[1]), record[2], record[3], helper.MustParseMonth(record[4]), helper.MustParseInt(record[5]))
		customerChan <- customer
		wg.Add(1)
		go perform(ctx, usecase, &donationByCustomer, &countDonation)
	}
	helper.Clear(&record)
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
		//ถ้าเจอ error ให้ลดความเร็วลง 10 ms
		if strings.Contains(err.Error(), "API rate limit has been exceeded") {
			rateMultiple += 10
			rateLimit = time.Tick(time.Duration(rateMultiple) * time.Millisecond)
		}
		(countDonation.failDonation) += c.AmountSubunits
		log.Errorf("Failed when create transaction with customer %v get error %v", c.Name, err)
	} else {
		//ถ้าไม่เจอ error เลยลองเพิ่มความเร็ว 10 ms
		if rateMultiple-10 >= 0 {
			rateMultiple -= 10
			rateLimit = time.Tick(time.Duration(rateMultiple) * time.Millisecond)
		}
		v, have := (*donation)[c.Name]
		if !have {
			(*donation)[c.Name] = c.AmountSubunits
		} else {
			(*donation)[c.Name] = v + c.AmountSubunits
		}
		(countDonation.sucessDonation) += c.AmountSubunits
	}
	(countDonation.totalDonation) += c.AmountSubunits
	//Clear memmory of customer
	helper.Clear(&c)
}

func display(countDonation *CountDonation, donationByCustomer map[string]int64) {
	//Display top donation
	fmt.Println("done.")
	fmt.Println("total received:", money.New(countDonation.totalDonation, "THB").Display())
	fmt.Println("successfully donated:", money.New(countDonation.sucessDonation, "THB").Display())
	fmt.Println("faulty donation:", money.New(countDonation.failDonation, "THB").Display())
	if len(donationByCustomer) == 0 {
		return
	}
	fmt.Println("average per person: THB", money.New(countDonation.sucessDonation/int64(len(donationByCustomer)), "THB").Display())

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
