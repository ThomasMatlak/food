package controller

import "context"

func HandleRequest(work func(context.Context)) {
	ctx := context.Background()
	work(ctx)
}
