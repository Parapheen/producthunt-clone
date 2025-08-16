package humantime

import (
	"strconv"
	"time"
)

func HumanTime(ts time.Time) string {
	d := time.Since(ts)
	if d < time.Minute {
		return "только что"
	}
	if d < time.Hour {
		return strconv.Itoa(int(d.Minutes())) + " мин назад"
	}
	if d < 24*time.Hour {
		return strconv.Itoa(int(d.Hours())) + " ч назад"
	}
	days := int(d.Hours() / 24)
	if days < 30 {
		return strconv.Itoa(days) + " дн назад"
	}
	months := days / 30
	if months < 12 {
		return strconv.Itoa(months) + " мес назад"
	}
	years := months / 12
	return strconv.Itoa(years) + " г назад"
}
