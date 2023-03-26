package model

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/ThomasMatlak/food/util"
)

type Resource struct {
	Created      *time.Time `json:"created"`
	LastModified *time.Time `json:"last_modified"`
	Deleted      *time.Time `json:"deleted"`
}

func ResourceId(labels []string) (string, error) {
	filteredLabels := util.FilterArray(labels, func(s string) bool { return len(s) > 0 })
	if len(filteredLabels) == 0 {
		return "", errors.New("tried to generate a resource id with an empty list of labels")
	}

	sort.Strings(filteredLabels)
	lowerCasedLabels := util.MapArray(filteredLabels, func(s string) string { return strings.ToLower(s) })
	return fmt.Sprintf("grn:tm-food:%s:%s", strings.Join(lowerCasedLabels, ":"), generateRandomString(10)), nil
}

func generateRandomString(length int) string { // thanks, chatgpt
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create a buffer to hold the generated string
	buffer := make([]byte, length)

	// Generate a random string by selecting a random character from the charset
	for i := range buffer {
		buffer[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(buffer)
}
