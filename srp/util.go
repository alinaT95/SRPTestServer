package srp

import (
	"bufio"
	"fmt"
	"math/big"
	"strings"
)

// NumberFromString converts a string (hex) to a number
func NumberFromString(s string) *big.Int {
	n := strings.Replace(s, " ", "", -1)

	result := new(big.Int)
	result.SetString(strings.TrimPrefix(n, "0x"), 16)

	return result
}

// max of integer arguments
// (because go doesn't give me "a > b ? a : b" )
func maxInt(n1 int, nums ...int) int {
	max := n1
	for _, n := range nums {
		if n > max {
			max = n
		}
	}
	return max
}

// bigIntFromBytes converts a byte array to a number
func bigIntFromBytes(bytes []byte) *big.Int {
	result := new(big.Int)
	for _, b := range bytes {
		result.Lsh(result, 8)
		result.Add(result, big.NewInt(int64(b)))
	}
	return result
}

func RemoveQuotesFromJson(json string) string {
	var flag = true
	var jsonTransformed = json
	for {
		var index = strings.Index(jsonTransformed, "x\":\"")

		var stringToReplace = "x\":\""
		if (index == -1) {

			index = strings.Index(jsonTransformed, "y\":\"")
			stringToReplace = "y\":\""
			if (index == -1) {
				flag = false
			}
		}
		if !flag {
			break
		}
		jsonTransformed = removeQuote(jsonTransformed, index)
		jsonTransformed = strings.Replace(jsonTransformed, stringToReplace, stringToReplace[0:len(stringToReplace) - 1], 1)
	}
	return jsonTransformed
}

func removeQuote(json string, startInd int) string{
	var indOfQuote = strings.Index(json[startInd+4: len(json)], "\"")
	return json[0: startInd + 4 + indOfQuote] +  json[startInd + 4 + indOfQuote + 1: len(json)]
}

func Write(w *bufio.Writer, data []byte) error {
	fmt.Printf("> %s\n", string(data))
	w.Write(data)
	w.Write([]byte("\n"))
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func Read(r *bufio.Reader) ([]byte, error) {
	fmt.Print("< ")
	data, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	fmt.Print(string(data))
	return data[:len(data)-1], nil
}


/**
 ** Copyright 2017 AgileBits, Inc.
 ** Licensed under the Apache License, Version 2.0 (the "License").
 **/
