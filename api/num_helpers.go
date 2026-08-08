package api

import "strconv"

// intStr formats an int as a string.
func intStr(n int) string { return strconv.Itoa(n) }

// int64Str formats an int64 as a string.
func int64Str(n int64) string { return strconv.FormatInt(n, 10) }

// floatStr formats a float64 as a string.
func floatStr(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }
