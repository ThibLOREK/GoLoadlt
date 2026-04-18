package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Next calcule la prochaine occurrence d'une expression cron standard (5 champs)
// à partir de t. Supporte * / , - sur les 5 champs.
// Pour la production, remplacer par github.com/robfig/cron/v3.
func Next(expr string, from time.Time) (time.Time, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return time.Time{}, fmt.Errorf("cron: expected 5 fields, got %d", len(fields))
	}

	// Avancer d'au moins une minute
	t := from.Add(time.Minute).Truncate(time.Minute)

	for i := 0; i < 366*24*60; i++ {
		if matchField(fields[3], int(t.Month()), 1, 12) &&
			matchField(fields[4], int(t.Weekday()), 0, 6) &&
			matchField(fields[2], t.Day(), 1, 31) &&
			matchField(fields[1], t.Hour(), 0, 23) &&
			matchField(fields[0], t.Minute(), 0, 59) {
			return t, nil
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}, fmt.Errorf("cron: no next occurrence found within 1 year")
}

func matchField(field string, value, min, max int) bool {
	if field == "*" {
		return true
	}
	for _, part := range strings.Split(field, ",") {
		if strings.Contains(part, "/") {
			sub := strings.SplitN(part, "/", 2)
			step, err := strconv.Atoi(sub[1])
			if err != nil || step <= 0 {
				continue
			}
			base := min
			if sub[0] != "*" {
				b, err := strconv.Atoi(sub[0])
				if err != nil {
					continue
				}
				base = b
			}
			for v := base; v <= max; v += step {
				if v == value {
					return true
				}
			}
		} else if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			lo, err1 := strconv.Atoi(bounds[0])
			hi, err2 := strconv.Atoi(bounds[1])
			if err1 == nil && err2 == nil && value >= lo && value <= hi {
				return true
			}
		} else {
			v, err := strconv.Atoi(part)
			if err == nil && v == value {
				return true
			}
		}
	}
	return false
}
