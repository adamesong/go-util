package image

import (
	"path/filepath"
	"regexp"
	"strings"
)

// GetThumbnailName 通过原文件名，获得缩略图的文件名。
// 例如：1.jpg -> 1_thumb.jpg
// 例如：abc.com/1.jpg -> abc.com/1_thumb.jpg
// 例如：https://abc.com/1.jpg -> https://abc.com/1_thumb.jpg
func GetThumbnailName(fileName string) (thumbName string) {
	if fileName != "" {
		// 获得文件的扩展名
		fileNameExt := filepath.Ext(fileName)                                // 例如.png
		fileNameWithoutExt := strings.Replace(fileName, fileNameExt, "", -1) // 获得不带扩展名的文件名
		thumbName = fileNameWithoutExt + "_thumb" + fileNameExt
		return
	}
	return
}

// GetOriginalName 通过thumbnail文件名，算出原图的文件名。如果本身不是thumbnail文件，则返回空字符串""。
// 例如：1_thumb.jpg -> 1.jpg
// 例如：abc.com/1_thumb.jpg -> abc.com/1.jpg
// 例如：https://abc.com/1_thumb.jpg -> https://abc.com/1.jpg
// 例如：https://abc.com/1_thumb -> https://abc.com/1
// 例如：https://abc.com/1.jpg -> ""
func GetOriginalName(thumbName string) (originName string) {
	if thumbName != "" {
		thumbExt := filepath.Ext(thumbName)

		var re *regexp.Regexp
		if thumbExt != "" {
			re = regexp.MustCompile(`^(.*)(_thumb)` + `\` + thumbExt + `$`) // 查找不包含"/"的、结尾是文件扩展名的字符串
		} else {
			re = regexp.MustCompile(`(^.*)(_thumb)$`) // 如果没有扩展名，_thumb后面没有东西了
		}
		result := re.FindStringSubmatch(thumbName) // 在base name中
		if len(result) > 0 {
			return result[1] + thumbExt
		}
	}
	return
}

// 与上面GetThumbnailName重复了
func GetThumbURL(imageURL string) string {
	if imageURL != "" {
		imageExt := filepath.Ext(imageURL)
		imageBase := filepath.Base(imageURL)
		imageDir := strings.Replace(imageURL, imageBase, "", 1)
		imageBaseWithoutExt := strings.Replace(imageBase, imageExt, "", -1)
		thumbURL := imageDir + imageBaseWithoutExt + "_thumb" + imageExt
		return thumbURL
	}
	return ""
}

// 通过原文件名的路径，获得缩略图的BaseName和路径
// 例如：abc.com/1.jpg -> abc.com/, 1_thumb.jpg
// 例如：https://abc.com/1.jpg -> https://abc.com/, 1_thumb.jpg
func GenerateThumbBaseName(fileNameWithPath string) (path, thumbName string) {
	if fileNameWithPath != "" {
		fileExt := filepath.Ext(fileNameWithPath)
		fileBase := filepath.Base(fileNameWithPath)
		path = strings.Replace(fileNameWithPath, fileBase, "", 1)
		fileBaseWithoutExt := strings.Replace(fileBase, fileExt, "", -1)
		thumbName = fileBaseWithoutExt + "_thumb" + fileExt
	}
	return
}

// 通过原文件名，获得s3或oss的bucket下的路径和文件名。
// 例如：http://cdn.aaa.com/dir/abc.png -> dir/abc.png
// 例如：https://oss.aliyun.com/dir/abc.png -> dir/abc.png
// 例如：/some/dir/abc.png -> some/dir/abc.png  多余的"/"也会被去除
// 例如：some/dir/abc.png -> some/dir/abc.png
func GetPathFileName(fileName string) (pathFileName string) {
	if fileName != "" {
		// 查找 http://xxx/ 或https://xxx/ 或 / (开头为/）
		re := regexp.MustCompile(`^http://[^/]*/?|^https://[^/]*/?|^/`)
		pathFileName = re.ReplaceAllString(fileName, "")
		return
	} else {
		return ""
	}
}

// 通过原文件名，获得s3或oss的bucket下的缩略图的路径和文件名。
// 例如：http://cdn.aaa.com/dir/abc.png -> dir/abc_thumb.png
// 例如：https://oss.aliyun.com/dir/abc.png -> dir/abc_thumb.png
// 例如：/some/dir/abc.png -> some/dir/abc_thumb.png  多余的"/"也会被去除
// 例如：some/dir/abc.png -> some/dir/abc_thumb.png
func GeneratePrefixThumbName(fileName string) (prefixThumbName string) {
	if fileName != "" {
		thumbName := GetThumbnailName(fileName)
		prefixThumbName = GetPathFileName(thumbName)
		return
	} else {
		return ""
	}
}

// 判断一个文件名是否符合缩略图文件名
// 例如：IsThumbFileName("1_thumb.jpg") true
//	    IsThumbFileName("http://www.com/1_thumb") true
//	    IsThumbFileName("abc_thumb11.jpg") false
func IsThumbFileName(fileName string) bool {
	if fileName != "" {
		fileExt := filepath.Ext(fileName)
		fileBase := filepath.Base(fileName)

		var re *regexp.Regexp
		if fileExt != "" {
			re = regexp.MustCompile(`([^/]*_thumb)` + `\` + fileExt + `$`) // 查找不包含"/"的、结尾是文件扩展名的字符串
		} else {
			re = regexp.MustCompile(`([^/]*_thumb)$`) // 如果没有扩展名，_thumb后面没有东西了
		}
		result := re.FindStringSubmatch(fileBase) // 在base name中
		if len(result) > 0 {
			return true
		}
	}
	return false
}

// 在含有完整路径的文件名中寻找如下字段，merchant_img comment_img trend_img profile_photo，只要找到则立刻返回。否则返回空字符串""
// 返回值 trend | comment | merchant | user
func GetImgCategory(fileNameWithPath string) string {

	if matched, _ := regexp.MatchString(`/merchant_img/`, fileNameWithPath); matched {
		return "merchant_img"
	} else if matched, _ := regexp.MatchString(`/trend_img/`, fileNameWithPath); matched {
		return "trend_img"
	} else if matched, _ := regexp.MatchString(`/comment_img/`, fileNameWithPath); matched {
		return "comment_img"
	} else if matched, _ := regexp.MatchString(`/profile_photo/`, fileNameWithPath); matched {
		return "profile_photo"
	} else {
		return ""
	}
}
