package day

import (
	"fmt"
	"net/http"
	"time"
)

// Display: Happy <day of the week>.
func GetHandler(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	fmt.Fprintf(w, "Happy %s.\n", t.Local().Weekday())
}
