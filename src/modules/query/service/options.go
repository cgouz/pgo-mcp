package service

import "time"

type Options struct {
	Timeout  time.Duration
	MaxRows  int
	MaxBytes int
}
