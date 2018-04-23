package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"./cipher"
)

//go:generate go build -o gen ./generator
//go:generate ./gen ./data/fng.1000.csv
/* Model */
type donator struct {
	name     string
	donation int // not thread-safe
}

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
				}
				if newdonator.donation, err = strconv.Atoi(row[1]); err != nil {
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

	var total int
	for _, donator := range donators {
		total += donator.donation
		s := fmt.Sprintf("Donate by %s amount %d", donator.name, donator.donation)
		fmt.Println(s)
	}
	s := fmt.Sprintf("total received: THB  %d", total)
	fmt.Println(s)
}
