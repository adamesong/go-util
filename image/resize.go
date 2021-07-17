package image

// 这个包很牛：https://github.com/esimov/caire
// 看着不错：https://github.com/willnorris/imageproxy
// 简单好用：https://github.com/disintegration/imaging

import (
	"bytes"
	"image"
	"io"

	"github.com/disintegration/imaging"
)

// resizeImage 给图片重新设置尺寸。如想按原比例，则仅提供一个宽度或一个高度。
// 从request中读取文件，然后resize，然后写入buffer
// filename: 是上传的文件的名字，不含路径，用于通过扩展名判断文件格式，以及命名新缩略图
// mustWidth: true: 不管原图大小，都按照此宽度缩放；false: 如果原图小于此宽度，则缩略图按原图尺寸生成
// mustHeight: true：不管原图大小，都按照此高度缩放；false: 如果原图小于此宽度，则缩略图按原图尺寸生成
// mustWidth和mustHeight一个true一个false，则按照true的那个来
// mustWidth和mustHeight两个都是true，则按照此长款生成，会拉伸，如想按原比例，则仅提供一个宽度或一个高度。
func ResizeImage(r io.Reader, filename string, width int, height int, mustWidth, mustHeight bool) (buff bytes.Buffer,
	err error) {
	srcImg, err := imaging.Decode(r, imaging.AutoOrientation(true)) //imaging.Open()  // 从文件路径读取文件
	if err != nil {
		return
	}
	imgFormat, err := imaging.FormatFromFilename(filename) // 获得图片的格式format
	if err != nil {
		return
	}
	srcWidth := srcImg.Bounds().Dx()  // 原图宽度
	srcHeight := srcImg.Bounds().Dy() // 原图高度
	var dstImg *image.NRGBA
	if mustWidth && mustHeight {
		// imaging.NearestNeighbor, .Linear, .CatmullRom, .Lanczos From faster (lower quality) to slower (higher quality)
		// If one of width or height is 0, the image aspect ratio is preserved.
		dstImg = imaging.Resize(srcImg, width, height, imaging.Lanczos)
	} else if mustWidth {
		dstImg = imaging.Resize(srcImg, width, 0, imaging.Lanczos)
	} else if mustHeight {
		dstImg = imaging.Resize(srcImg, 0, height, imaging.Lanczos)
	} else { // 如果都不是必须缩小或放大至指定尺寸，则判断原图是不是比该尺寸小；
		if srcWidth <= width || srcHeight <= height { // 只要原图的长或宽小于指定的长或宽，就不变
			dstImg = imaging.Resize(srcImg, srcWidth, srcHeight, imaging.Lanczos)
		} else { // 如果原图的长和宽都大于指定的长和宽，则按指定尺寸缩小
			dstImg = imaging.Resize(srcImg, width, height, imaging.Lanczos)
		}
	}
	if err = imaging.Encode(&buff, dstImg, imgFormat); err != nil {
		return
	}
	return
}
