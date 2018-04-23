package main

import (
	"context"
	"encoding/csv"
	"io"
)

func process(in io.Reader, out io.Writer) error {
	pipe := NewPipeline(context.Background())
	go importRecords(pipe, in)
	go exportRecords(pipe, out)

	pipe.Wait()
	return pipe.err
}

func importRecords(pipe *Pipeline, r io.Reader) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = 9
	reader.ReuseRecord = true
	if _, err := reader.Read(); err != nil { // header row
		pipe.CancelWithError(err)
		return
	}

	defer pipe.Close()

	for {
		record := &Record{}
		if row, err := reader.Read(); err != nil {
			if err == io.EOF {
				return
			} else {
				pipe.CancelWithError(err)
				return
			}
		} else {
			record.ParseCSV(row)
		}

		if !pipe.Send(record) {
			return
		}
	}
}

func exportRecords(pipe *Pipeline, w io.Writer) {
	var record *Record

	writer := csv.NewWriter(w)
	if err := writer.Write(record.CSVHeader()); err != nil {
		pipe.CancelWithError(err)
	}

	defer writer.Flush()
	for record = range pipe.Recv() {
		writer.Write(record.CSV())
	}
}
