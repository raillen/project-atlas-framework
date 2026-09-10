//go:build ignore

package main

import "context"

func Fetch(ctx context.Context) error {
	return ctx.Err()
}
