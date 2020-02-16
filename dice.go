package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

var validDice []int = []int{4, 6, 8, 10, 12, 20, 100}

func getRoll(ceiling int) int {
	// [0,n] exclusive on the upper bound
	out := rand.Intn(ceiling)
	return out + 1
}

func parseDice(s string) (string, error) {
	split := strings.Split(s, "d")
	plus := 0

	plusRaw := strings.Split(split[1], "+")
	if plusRaw[0] == "69" {
		return "n i c e", nil
	}

	plus, err := strconv.Atoi(plusRaw[len(plusRaw)-1])
	if err != nil || len(plusRaw) == 1 {
		plus = 0
	}

	diceNum, err := strconv.Atoi(split[0])
	if err != nil {
		return "", errors.New("unable to parse dice quantity")
	}
	if diceNum > 100 {
		return "", errors.New("too many dice jfc")
	}

	diceCeiling, err := strconv.Atoi(plusRaw[0])
	if err != nil {
		return "", errors.New("unable to parse die type")
	}

	valid := false
	for _, kind := range validDice {
		if diceCeiling == kind {
			valid = true
			break
		}
	}
	if !valid {
		return "", errors.New("invalid dice type")
	}

	out := ""
	var totes []int
	for i := 0; i < diceNum; i++ {
		res := getRoll(diceCeiling)
		totes = append(totes, res)
		if i == diceNum-1 {
			out += fmt.Sprintf("%d", res)
		} else {
			out += fmt.Sprintf("%d  ", res)
		}
	}
	total := 0
	for _, d := range totes {
		total += d
	}
	out += fmt.Sprintf(",  total: %d", total+plus)

	return out, nil
}
