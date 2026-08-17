package api

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"time"
)

func framebufferPNG(fb []byte) ([]byte, error) {
	if len(fb) < 4096 {
		padded := make([]byte, 4096)
		copy(padded, fb)
		fb = padded
	}
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			c := fb[y*64+x]
			img.Set(x, y, rgb332ToRGBA(c))
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func rgb332ToRGBA(c byte) color.RGBA {
	r := (c >> 5) & 0x7
	g := (c >> 2) & 0x7
	b := c & 0x3
	return color.RGBA{
		R: uint8((r * 255) / 7),
		G: uint8((g * 255) / 7),
		B: uint8((b * 255) / 3),
		A: 255,
	}
}

func requestContext(r interface {
	Context() context.Context
}, fallback time.Duration) (context.Context, context.CancelFunc) {
	ctx := r.Context()
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, fallback)
}
