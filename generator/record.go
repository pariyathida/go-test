package main

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	name   string
	amount int64

	ccNumber          string
	ccCVV             string
	ccExpirationMonth int
	ccExpirationYear  int

	err error
}

func (record *Record) ParseCSV(data []string) {
	// Number,CCType,CCNumber,CVV2,CCExpires,Title,GivenName,MiddleInitial,Surname
	// 0     ,1     ,2       ,3   ,4        ,5    ,6        ,7            ,8

	record.name = strings.Join(data[5:], " ")
	record.amount = 100000 + rand.Int63n(5000000)

	record.ccNumber = data[2]
	record.ccCVV = data[3]

	record.ccExpirationMonth = int(time.Now().Month())
	record.ccExpirationYear = time.Now().Year() + 1
	if idx := strings.IndexRune(data[4], '/'); idx > 0 {
		if expMonth, err := strconv.Atoi(data[4][:idx]); err == nil {
			record.ccExpirationMonth = expMonth
		}
		if expYear, err := strconv.Atoi(data[4][idx+1:]); err == nil {
			record.ccExpirationYear = expYear
		}
	}
}

func (record *Record) CSVHeader() []string {
	return []string{
		"Name",
		"AmountSubunits",
		"CCNumber",
		"CVV",
		"ExpMonth",
		"ExpYear",
	}
}

func (record *Record) CSV() []string {
	if record == nil {
		return []string{"", "", "", "", "", ""}
	}

	fmt := func(n int64) string { return strconv.FormatInt(n, 10) }

	return []string{
		record.name,
		fmt(record.amount),
		record.ccNumber,
		record.ccCVV,
		fmt(int64(record.ccExpirationMonth)),
		fmt(int64(record.ccExpirationYear)),
	}
}
