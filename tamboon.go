package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
	"sort" 
	
	"./cipher"
    "github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
)

//go:generate go build -o gen ./generator
//go:generate ./gen ./data/fng.1000.csv
/* Model */
type donator struct {
	name     string
	donation int64 // not thread-safe
	ccNumber string
	cvv      string
	expMonth int
	expYear  int
}

/* Variables for Summary */
var totalReceived int64
var successDonated int64
var donatorNumber int

/* Reader & Writer */
func writeFile(w io.Writer, p []byte) (int, error) {
	buffer := make([]byte, 4096, 4096)
	n := copy(buffer, p)

	return w.Write(buffer[:n])
}

func readEncryptedFile(filename string) []byte {
	/* Open File */
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	/* Create Reader */
	reader, err := cipher.NewRot128Reader(file)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 4096, 4096)
	reader.Read(buf)
	if err != nil {
		panic(err)
	}
	return buf
}

func readDecodedFile(filename string) (donators []donator) {
	f, err := os.Open(filename)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	csvr := csv.NewReader(f)
	if err != nil {
		log.Panic(err)
	}

	var totalFields int
	i := 0
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				return donators
			}
		}

		if i != 0 {
			if len(row) == totalFields {
				newdonator := donator{
					name: row[0],
					ccNumber: row[2],
					cvv: row[3],
				}
				if newdonator.donation, err = strconv.ParseInt(row[1], 10, 64); err != nil {
					log.Panic(err)
				}
				if newdonator.expMonth, err = strconv.Atoi(row[4]); err != nil {
					log.Panic(err)					
				}
				if newdonator.expYear, err = strconv.Atoi(row[5]); err != nil {
					log.Panic(err)					
				}
				donators = append(donators, newdonator)
			}
		} else {
			totalFields = len(row)
		}
		i++
	}
}

/*Charge*/
const (
	/* Read these from environment variables or configuration files! */
	OmisePublicKey = "pkey_test_5bpa0b3cpyakzssqckg"
	OmiseSecretKey = "skey_test_5bp75648opp5swpa6kk"
)

func charge(donator donator){
	client, e := omise.NewClient(OmisePublicKey, OmiseSecretKey)
	if e != nil {
		log.Fatal(e)
	}
	
	expMonth := time.Month(donator.expMonth)

	/* Creates a token from a test card.*/
	token, createToken := &omise.Token{}, &operations.CreateToken{
		Name:            donator.name,
		Number:          donator.ccNumber,
		ExpirationMonth: expMonth,
		ExpirationYear:  donator.expYear,
	}
	if e := client.Do(token, createToken); e != nil {
		log.Fatal(e)
	}

	/* Creates a charge from the token */
	charge, createCharge := &omise.Charge{}, &operations.CreateCharge{
		Amount:   donator.donation,
		Currency: "thb",
		Card:     token.ID,
	}
	if e := client.Do(charge, createCharge); e != nil {
		log.Fatal(e)
	}
    
	successDonated += donator.donation
	donatorNumber += 1
	//log.Printf("charge: %s  amount: %s %d\n", charge.ID, charge.Currency, charge.Amount)
}

/* sort */
type By func(d1, d2 *donator) bool

type donatorSorter struct {
	donators []donator
	by func(d1, d2 *donator) bool 
}

func (by By) Sort(donators []donator) {
	ds := &donatorSorter{
		donators: donators,
		by:      by, 
	}
	sort.Sort(ds)
}

func (s *donatorSorter) Len() int {
	return len(s.donators)
}

func (s *donatorSorter) Swap(i, j int) {
	s.donators[i], s.donators[j] = s.donators[j], s.donators[i]
}

func (s *donatorSorter) Less(i, j int) bool {
	return s.by(&s.donators[i], &s.donators[j])
}


/* Main function */
func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s (file)\n", os.Args[0])
	}

	var (
		fngName     = os.Args[1]
		encodedName = fngName + ".rot128"
	)

	/* Create Decoded File */
	outFile, err := os.Create(fngName)
	if err != nil {
		log.Fatalln(err)
	}
	buf := readEncryptedFile(encodedName)
	writeFile(outFile, buf)

	donators := readDecodedFile(fngName)

	/* perform donations */
	fmt.Println("performing donations...")	
	for _, donator := range donators {
		totalReceived += donator.donation	
		//s := fmt.Sprintf("Donate by %s amount %d cardnum %s %s %d %d", donator.name, donator.donation, donator.ccNumber, donator.cvv, donator.expMonth, donator.expYear)
		//fmt.Println(s)
		charge(donator)
	}
	fmt.Print("done.\n\n")
	
	/* summary */
	s1 := fmt.Sprintf("      total received : THB  %.2f", float64(totalReceived)/100)
	fmt.Println(s1)
	s2 := fmt.Sprintf("successfully donated : THB  %.2f", float64(successDonated)/100)
	fmt.Println(s2)
	s3 := fmt.Sprintf("     faulty donation : THB  %.2f", float64(totalReceived-successDonated)/100)
	fmt.Println(s3)

	avgPerPerson := (float64(totalReceived)/float64(donatorNumber))/100
	s4 := fmt.Sprintf("\n  average per person : THB  %.2f", avgPerPerson)
	fmt.Println(s4)   
	
	/* sorting */
	donation := func(d1, d2 *donator) bool {
		return d1.donation > d2.donation
	}	
	By(donation).Sort(donators)
	//fmt.Println("By donation:", donators)

	/* display top donors */
	fmt.Print("          top donors : ")
	fmt.Print(donators[0].name)
	fmt.Print("\n                       ")
	fmt.Print(donators[1].name)
	fmt.Print("\n                       ")
	fmt.Print(donators[2].name)
}
