package gmq

import (
	"context"
	"time"

	"github.com/giant-stone/go/glogging"
	"github.com/giant-stone/go/gstr"
)

func LoggingElapsed(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, msg IMsg) error {
		start := time.Now()
		err := h.ProcessMsg(ctx, msg)
		if err != nil {
			shorten := gstr.ShortenWith(msg.String(), 100, gstr.DefaultShortenSuffix)
			glogging.Sugared.Warnf("ProcessMsg failed %v %s", time.Since(start), shorten)
			return err
		}

		glogging.Sugared.Debugf("ProcessMsg success %v %s", time.Since(start), msg.GetId())
		return nil
	})
}
