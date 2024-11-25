package image

import (
	"bytes"
	"context"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"image/png"
	"log/slog"
)

// Options are used in two cases: first as default parameters for all source sets which will be created
// and second deviating options for a specific case, where the default options shall not apply.
type Options struct {
	// Default is 32MiB
	MaxFileSize int64
	// Default is 3840 (4k)
	MaxWidthOrHeight int
}

// CreateSrcSet accepts the given files and returns at least a src set.
type CreateSrcSet func(user auth.Subject, customOpts Options, img core.File) (SrcSet, error)

func NewCreateSrcSet(opts Options, srcSets Repository, images blob.Store) CreateSrcSet {
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = 1024 * 1024 * 32
	}

	if opts.MaxWidthOrHeight == 0 {
		opts.MaxWidthOrHeight = 3840
	}

	return func(user auth.Subject, customOpts Options, img core.File) (SrcSet, error) {
		// inherit and overload default Options
		if customOpts.MaxWidthOrHeight == 0 {
			customOpts.MaxWidthOrHeight = opts.MaxWidthOrHeight
		}

		if customOpts.MaxFileSize == 0 {
			customOpts.MaxFileSize = opts.MaxFileSize
		}

		opts = customOpts

		if s, _ := img.Size(); s > opts.MaxFileSize {
			return SrcSet{}, std.NewLocalizedError("Datei ist zu groß", fmt.Sprintf("Das Bild '%s' ist %.2f MiB groß und überschreitet die maximal erlaubte Dateigröße von %.2f MiB.", img.Name(), float64(s)/1024/1024, float64(opts.MaxFileSize)/1024/1024))
		}

		// note, that our default blob store implementation applies automatic deduplication internally,
		// thus if even though we burn I/O bandwidth and CPU cycles, we don't have a problem with redundant
		// bytes on storage.
		srcImgId := data.RandIdent[ID]()

		// we use cancel to clean up broken writes into the blob store, if it comes too late, it is just ignored.
		ctx, cancel := context.WithCancel(context.Background())

		// just persist original image blob
		w, err := images.NewWriter(ctx, string(srcImgId))
		if err != nil {
			cancel()
			return SrcSet{}, fmt.Errorf("cannot open image writer: %w", err)
		}

		encodedNumSrcBytes, err := img.Transfer(w)
		if err != nil {
			cancel()
			_ = w.Close()
			return SrcSet{}, fmt.Errorf("cannot transfer image: %w", err)
		}

		// close first, this actually persists the data
		if err := w.Close(); err != nil {
			cancel()
			return SrcSet{}, fmt.Errorf("cannot commit image writer: %w", err)
		}

		// avoid ctx leak but do not cancel write, thus this must be after w.Close
		cancel()

		// let us check, if we got a (zip) bomb, e.g. a single colored pixel blown up
		// in gigantic dimensions, usually png but potentially jpg as well
		optR, err := images.NewReader(context.Background(), string(srcImgId))
		if err != nil {
			return SrcSet{}, fmt.Errorf("cannot open reader from just stored image: %w", err)
		}

		r := optR.Unwrap()
		imgCfg, _, err := image.DecodeConfig(r)
		_ = r.Close()

		if err != nil {
			return SrcSet{}, std.NewLocalizedError("Nicht unterstütztes Format", fmt.Sprintf("Das Bild '%s' kann nicht dekodiert werden. Es sind nur PNG und JPEG möglich.", img.Name()))
		}

		if imgCfg.Width > opts.MaxWidthOrHeight || imgCfg.Height > opts.MaxWidthOrHeight {
			return SrcSet{}, std.NewLocalizedError("Bilddimension zu groß", fmt.Sprintf("Das Bild '%s' darf nicht größer als %dx%d Pixel sein. Es hat aber %dx%d.", img.Name(), opts.MaxWidthOrHeight, opts.MaxWidthOrHeight, imgCfg.Width, imgCfg.Height))
		}

		// open again for a full decode run
		optR, err = images.NewReader(context.Background(), string(srcImgId))
		if err != nil {
			return SrcSet{}, fmt.Errorf("cannot open reader from just stored image: %w", err)
		}

		r = optR.Unwrap()
		src, imgTyp, err := image.Decode(r)
		_ = r.Close()
		if err != nil {
			return SrcSet{}, std.NewLocalizedError("Fehler bei der Dekodierung", fmt.Sprintf("Das Bild '%s' kann nicht dekodiert werden, da es möglicherweise defekt ist oder Features enthält, die wir nicht unterstützen.", img.Name()))
		}

		// now with the in-memory image, build our pyramid using 1/2 billinear sampling.
		// we first round down to a multiple of 8, so that we always optimize DCT matching
		// for JPEG compression, if not already done.
		srcSet := SrcSet{
			ID: srcImgId,
		}

		// put the original image for completeness
		srcSet.Images = append(srcSet.Images, Image{
			Width:  imgCfg.Width,
			Height: imgCfg.Height,
			Data:   srcImgId,
		})

		thumbWidth, thumbHeight := src.Bounds().Max.X, src.Bounds().Max.Y
		tmp := bytes.NewBuffer(make([]byte, 0, encodedNumSrcBytes)) // pre-alloc, probably not larger than source (well depends...)
		for {
			thumbWidth, thumbHeight = nextLowerMultipleOfEight(thumbWidth/2), nextLowerMultipleOfEight(thumbHeight/2)
			if thumbWidth < 32 || thumbHeight < 32 {
				break
			}

			dst := image.NewRGBA(image.Rect(0, 0, thumbWidth, thumbHeight))

			// Resize:
			draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

			// reuse our in-mem buffer
			tmp.Reset()

			// encode what is native to our format, e.g. to preserve quality, compression-characteristics
			// and transparency
			switch imgTyp {
			case "png":
				if err := png.Encode(tmp, dst); err != nil {
					return SrcSet{}, fmt.Errorf("png image encode error: %w", err)
				}
				srcSet.Format = FormatPng
			default:
				if err := jpeg.Encode(tmp, dst, &jpeg.Options{Quality: 85}); err != nil {
					return SrcSet{}, fmt.Errorf("jpeg image encode error: %w", err)
				}
				srcSet.Format = FormatJpeg
			}

			thumbImgId := data.RandIdent[ID]()
			if err := blob.Put(images, string(thumbImgId), tmp.Bytes()); err != nil {
				return SrcSet{}, fmt.Errorf("image thumb put error: %w", err)
			}

			srcSet.Images = append(srcSet.Images, Image{
				Width:  dst.Bounds().Dx(),
				Height: dst.Bounds().Dy(),
				Data:   thumbImgId,
			})

			slog.Info("created image step", "src", srcImgId, "thumb", thumbImgId, "w", dst.Bounds().Dx(), "h", dst.Bounds().Dy())
		}

		// we are done, persist the entire calculated source set pyramid
		if err := srcSets.Save(srcSet); err != nil {
			return SrcSet{}, fmt.Errorf("save SrcSet error: %w", err)
		}

		return srcSet, nil
	}
}

func nextLowerMultipleOfEight(n int) int {
	if n < 8 {
		return 0
	}
	return (n / 8) * 8
}
