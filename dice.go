package main

import (
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

	diceNum, err := strconv.Atoi(split[0])
	if err != nil {
		return "", fmt.Errorf("unable to parse dice quantity")
	}

	diceCeiling, err := strconv.Atoi(split[1])
	if err != nil {
		return "", fmt.Errorf("unable to parse die type")
	}

	valid := false
	for _, kind := range validDice {
		if diceCeiling == kind {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("invalid dice type")
	}

	out := ""
	var totes []int
	for i := 0; i < diceNum; i++ {
		res := getRoll(diceCeiling)
		totes = append(totes, res)
		out += fmt.Sprintf("%d  ", res)
	}
	out = strings.TrimSpace(out)
	total := 0
	for _, d := range totes {
		total += d
	}
	out += fmt.Sprintf(", total: %d", total)

	return strings.TrimSpace(out), nil
}
