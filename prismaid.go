package prismaid

import (
	"github.com/open-and-sustainable/prismaid/review"
)

func RunReview(tomlConfiguration string) error {
	return review.RunReview(tomlConfiguration)
}
