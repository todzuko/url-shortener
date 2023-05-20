package routes

import "time"

type request struct {
	URL   string        `json:"url"`
	Short string        `json:"short"`
	Exp   time.Duration `json:"exp"`
}

type response struct {
	URL             string        `json:"url"`
	Short           string        `json:"short"`
	Exp             time.Duration `json:"exp"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ResolveUrl() {}
