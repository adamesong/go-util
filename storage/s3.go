package storage

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/adamesong/go-util/image"
	"github.com/adamesong/go-util/logging"
	"github.com/adamesong/go-util/random"

	// ! aws-sdk-go deprecated, use aws-sdk-go-v2 instead
	// todo 将此包替换为 aws-sdk-go-v2
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3 struct {
	Region          string // ie: us-west-2
	AccessKeyID     string
	AccessSecretKey string
	DefaultACL      string // ie: public-read
	BucketName      string // ie: xx-debug
	URL             string // ie: https:xx-debug.s3.us-west-2.amazonaws.com/, https://cdn.xx.com
	CDNHostName     string // ie: cdn.xxx.com

}

func (s3 *S3) getSession() *session.Session {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(s3.Region),
		Credentials: credentials.NewStaticCredentials(s3.AccessKeyID, s3.AccessSecretKey, ""),
	}))
	return sess
}

// AddFileToS3 上传文件到S3
// fileName 是文件名
// backetDir 是上传到bucket里的哪个文件夹，例如 "abc/upload"
// https://golangcode.com/uploading-a-file-to-s3/
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/
// https://docs.aws.amazon.com/zh_cn/sdk-for-go/v1/developer-guide/configuring-sdk.html
// https://stackoverflow.com/questions/48221701/how-to-upload-file-to-amazon-s3-using-gin-framework
func (s3 *S3) AddFileToS3(file io.Reader, fileName, bucketDir string) (url string, err error) {
	// 打开本地文件
	//f, err := os.Open(fullFileName)
	//if err != nil {
	//	return "", fmt.Errorf("failed to open file %q, %v", fullFileName, err)
	//}
	//defer func() {
	//	if err := f.Close(); err != nil {
	//		log.Fatal(err.Error())
	//	}
	//}()

	// Upload the file to S3.
	uploader := s3manager.NewUploader(s3.getSession())
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3.BucketName),
		//Key:    aws.String(bucketDir + "/" + path.Base(fullFileName)), // path.Base获得 xx.jpg这样的文件名，不带目录
		Key:  aws.String(bucketDir + "/" + fileName),
		ACL:  aws.String(s3.DefaultACL),
		Body: file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}

	//fmt.Printf("file uploaded to, %s\n", aws.StringValue(&result.Location))
	return aws.StringValue(&result.Location), nil
}

// AddImageToS3 在上传图片到S3的同时，生成指定尺寸的缩略图并上传到S3中的相同目录下。
func (s3 *S3) AddImageToS3(file io.Reader, fileName, bucketDir string, width, height int) (url string, err error) {
	// 上传原文件
	keyPrefix := random.RandomString(8) // 给文件名前面增加一个8位的随机字符串，以防止同名文件上传，导致覆盖。
	url, err = s3.AddFileToS3(file, keyPrefix+"_"+fileName, bucketDir)
	if err != nil {
		return
	}
	go func() {
		// 生成缩略图
		buff, err := image.ResizeImage(file, fileName, width, height, false, false)
		if err != nil {
			return
		}

		// 获得上传到s3之后的新文件名 上传的文件如果重名，可能会出现文件名后增加字符的情况，所以不能完全依赖老的fileName
		// 这种方式中文字符编码有问题 维多利亚.jpg -> %E7%BB%B4%E5%A4%9A%E5%88%A9%E4%BA%9A_thumb.jpg
		// 需要这样处理：params, err := url.ParseQuery(queryStr)
		//re := regexp.MustCompile(`([^/]*)` + `\` + fileNameExt + `$`) // 查找不包含"/"的、结尾是文件扩展名的字符串
		//result := re.FindStringSubmatch(url)
		//newFileNameWithoutExt := result[1]
		// 为了同django_ckeditor在运营后台上传图片自动生成的缩略图命名统一，改为：原文件名_thumb.原扩展名
		//resizeFileName := newFileNameWithoutExt + "_thumb" + fileNameExt

		resizeFileName := image.GetThumbnailName(keyPrefix + "_" + fileName)
		// 上传缩略图
		_, _ = s3.AddFileToS3(&buff, resizeFileName, bucketDir)
	}()

	return
}

// GetS3FileNameFromURL 将s3 url转换为在s3中的存储路径和文件名
// ok: 如果成功转换了文件路径，则返回true，此时fileName为空字符串。
// 存在S3中的文件url:
// 正式服务器 https://cdn.xxxx.com/upload/2019/07/04/homepage-ads_CxcMbiF.jpg
// 测试服务器形式1 https://xxx-debug.s3-us-west-2.amazonaws.com/upload/vehicle/2/report/favicon.png
// 测试服务器形式2 https://xxx-debug.s3.us-west-2.amazonaws.com/upload/vehicle/2/report/favicon.png
// 测试服务器形式3 https://s3.us-west-2.amazonaws.com/xxx-debug/upload/vehicle/2/report/favicon.png
// 测试服务器形式4 https://s3-us-west-2.amazonaws.com/xxx-debug/upload/vehicle/2/report/favicon.png
func (s3 *S3) GetS3FileNameFromURL(fileURL string) (fileName string, ok bool) {
	if path := "https://" + s3.CDNHostName + "/"; strings.Contains(fileURL, path) {
		fileName = strings.Replace(fileURL, path, "", 1)
		ok = true
	} else if path := "https://" + s3.BucketName + ".s3-" + s3.Region + ".amazonaws.com/"; strings.Contains(fileURL, path) {
		fileName = strings.Replace(fileURL, path, "", 1)
		ok = true
	} else if path := "https://" + s3.BucketName + ".s3." + s3.Region + ".amazonaws.com/"; strings.Contains(fileURL, path) {
		fileName = strings.Replace(fileURL, path, "", 1)
		ok = true
	} else if path := "https://s3." + s3.Region + ".amazonaws.com/" + s3.BucketName + "/"; strings.Contains(fileURL, path) {
		fileName = strings.Replace(fileURL, path, "", 1)
		ok = true
	} else if path := "https://s3-" + s3.Region + ".amazonaws.com/" + s3.BucketName + "/"; strings.Contains(fileURL, path) {
		fileName = strings.Replace(fileURL, path, "", 1)
		ok = true
	} else {
		return
	}
	return
}

// DeleteFileFromS3 删除S3中的一个文件
func (s *S3) DeleteFileFromS3(fileURL string) {
	// 判断老照片是不是S3的相应bucket的，如果是，删除

	s3FileName, found := s.GetS3FileNameFromURL(fileURL)

	if found {
		svc := s3.New(s.getSession())
		input := &s3.DeleteObjectInput{
			Bucket: aws.String(s.BucketName),
			Key:    aws.String(s3FileName),
		}

		_, err := svc.DeleteObject(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return
		}
		//fmt.Println(result)
	}

}

// DeleteImageAndThumbFromS3 删除图片及缩略图（缩略图的命名为http://xxx/xxx_thumb.xx）
func (s3 *S3) DeleteImageAndThumbFromS3(imageURL string) {
	// 获得thumbnail图片的RUL
	thumbURL := image.GetThumbURL(imageURL)
	s3.DeleteFileFromS3(imageURL) // 删除原图
	s3.DeleteFileFromS3(thumbURL) // 删除缩略图
}

// 由CDN的url的文件转为s3 url的文件路径
func (s *S3) getS3objects(fileURLs []string) []*s3.ObjectIdentifier {
	objects := make([]*s3.ObjectIdentifier, 0)
	for _, fileURL := range fileURLs {
		if strings.Contains(fileURL, s.URL) {
			s3FileName := strings.Replace(fileURL, s.URL, "", 1) // 把url中的域名部分去掉
			objects = append(objects, &s3.ObjectIdentifier{Key: aws.String(s3FileName)})
		}
	}
	return objects
}

// DeleteFilesFromS3 批量删除S3中的文件 https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#S3.DeleteObjects
func (s *S3) DeleteFilesFromS3(fileURLs []string) {
	svc := s3.New(s.getSession())
	objects := s.getS3objects(fileURLs)

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(s.BucketName),
		Delete: &s3.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false),
		},
	}

	//result, err := svc.DeleteObjects(input)
	_, err := svc.DeleteObjects(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
	//fmt.Println(result)
}

func (s3 *S3) DeleteImagesAndThumbsFromS3(imageURLs []string) {
	// 获得图片链接和缩略图链接
	allURLs := make([]string, 0)
	for _, imageURL := range imageURLs {
		// 如果是正式服务器，该imageURL可能是cdn url(https://cdn.xxxx.com/...)，需要转换成S3的url。
		if s3.CDNHostName != "" {
			imageURL = s3.GetS3URLFromCDNURL(imageURL)
		}
		thumbURL := image.GetThumbURL(imageURL)
		allURLs = append(allURLs, imageURL, thumbURL)
	}
	// 执行删除
	s3.DeleteFilesFromS3(allURLs)
}

// GetCDNURL 如果是正式服务器，则将s3 URL替换为CDN的url
// s3 URL的形式有如下几种：
// 1. https://s3.us-west-2.amazonaws.com/bucketName/dir...
// 2. https://bucketName.s3.us-west-2.amazonaws.com/dir...
// 3. https://xxx-debug.s3-us-west-2.amazonaws.com/dir...
func (s3 *S3) GetCDNURLFromS3URL(s3URL string) (cdnURL string) {
	cdnURL = s3URL // 初始值
	if s3URL != "" {

		// 针对第一种形式的s3 URL，如 https://s3.us-west-2.amazonaws.com/bucketName/dir...
		// 用正则找出url中包含bucket name的地方，然后去除bucket name之前的aws的url，再将bucketName替换为cdnHostName
		re := regexp.MustCompile(`^https://(\S*)` + "(" + s3.BucketName + "/)")
		// FindSubmatch查找子匹配项 第一个匹配的是全部元素 第二个匹配的是第一个()里面的 第三个匹配的是第二个()里面的
		result := re.FindStringSubmatch(s3URL)
		// 如果匹配到了，则替换（否则不替换）
		if len(result) > 0 && result[0] != "" {
			cdnURL = strings.Replace(s3URL, result[1]+s3.BucketName, s3.CDNHostName, 1)
			return
		}

		// 针对第二种形式和第三种形式的s3 URL，
		// 如 https://bucketName.s3.us-west-2.amazonaws.com/dir...
		// 又如 如 https://xxx-debug.s3-us-west-2.amazonaws.com/dir...
		re2 := regexp.MustCompile(`^https://(` + s3.BucketName + `.s3\S*.com)`)
		// FindSubmatch查找子匹配项 第一个匹配的是全部元素 第二个匹配的是第一个()里面的 第三个匹配的是第二个()里面的
		result2 := re2.FindStringSubmatch(s3URL)
		// 如果匹配到了，则替换（否则不替换）
		if len(result2) > 0 && result2[0] != "" {
			cdnURL = strings.Replace(s3URL, result2[1], s3.CDNHostName, 1)
			return
		}
	}
	return
}

// GetS3URLFromCDNURL 如果是正式服务器，将图片的cdn URL替换为在S3的真实URL
// s3的真实url 如：https://xxx-debug.s3-us-west-2.amazonaws.com/dir...
// 又如：https://s3.us-west-2.amazonaws.com/backetName/user_upload/78/merchant_img/WechatIMG63.jpeg
// s3 hostname由如下部分构成："s3." + S3_REGION + ".amazonaws.com/"
func (s3 *S3) GetS3URLFromCDNURL(cdnURL string) (s3URL string) {
	s3URL = cdnURL // 初始值
	if cdnURL != "" {
		re := regexp.MustCompile(`^(https://)` + "(" + s3.CDNHostName + "/)")

		// FindSubmatch查找子匹配项 第一个匹配的是全部元素 第二个匹配的是第一个()里面的 第三个匹配的是第二个()里面的
		result := re.FindStringSubmatch(cdnURL)
		// 如果匹配到了
		if len(result) > 0 && result[0] != "" {
			newStr := s3.BucketName + ".s3-" + s3.Region + ".amazonaws.com/"
			s3URL = strings.Replace(cdnURL, result[2], newStr, 1)
		}
	}
	return
}

// 获得s3 bucket中的文件列表（每次最多1000个）
// bucketName: bucketName
// prefix: 例如 "user_upload/"
// startAfter：从哪个开始，例如"assets/slider/5.JPG"
// continuationToken：当*result.IsTruncated==true时，会有个nextContinuationToken，用这个来取后面的objects
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#S3.ListObjectsV2
// 返回值：result
//
//	 Contents: [
//	   {
//	    ETag: "\"787d14e0d82d995e3615e0900575f951\"",
//	    Key: "assets/slider/4.JPG",
//	    LastModified: 2019-07-19 21:52:13 +0000 UTC,
//	    Size: 90043,
//	    StorageClass: "STANDARD"
//	  },
//	  {
//	    ETag: "\"7506c19233ddd9a02e29465fc070f0ef\"",
//	    Key: "assets/slider/5.JPG",
//	    LastModified: 2019-07-19 21:52:13 +0000 UTC,
//	    Size: 108746,
//	    StorageClass: "STANDARD"
//	  },
//	  {
//	    ETag: "\"d41d8cd98f00b204e9800998ecf8427e\"",
//	    Key: "baby/",
//	    LastModified: 2019-06-12 17:41:06 +0000 UTC,
//	    Size: 0,
//	    StorageClass: "STANDARD"
//	  }
//	]
//	IsTruncated: true,  // 是否被截断了
//	KeyCount: 6,
//	MaxKeys: 6,
//	Name: "classtop-com-test",
//	NextContinuationToken: "16MrB83O08WqafJ8HqilMdg/iSUSLSJMwZ6on7UtsYC0YWg4lBZHFLQ==",
//	Prefix: "",
//	StartAfter: "assets/"
func (s *S3) ListObjectsFromS3(bucketName, prefix, startAfter, continuationToken string) (result *s3.ListObjectsV2Output) {
	svc := s3.New(s.getSession())
	var input *s3.ListObjectsV2Input
	if continuationToken == "" {
		input = &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucketName),
			MaxKeys:           aws.Int64(1000), // 最大只能1000个
			Prefix:            aws.String(prefix),
			StartAfter:        aws.String(startAfter),
			ContinuationToken: nil,
		}
	} else {
		input = &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucketName),
			MaxKeys:           aws.Int64(1000), // 最大只能1000个
			Prefix:            aws.String(prefix),
			StartAfter:        aws.String(startAfter),
			ContinuationToken: aws.String(continuationToken),
		}
	}

	result, err := svc.ListObjectsV2(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	return
}

// 与上面的方法不同的仅仅是：将结果append到list中。且如果被截断了，递归执行，直至全部object都被append到list中
func (s3 *S3) GetObjectsListFromS3(bucketName, prefix, startAfter, continuationToken string, objList *[]*s3.Object) {
	result := s3.ListObjectsFromS3(bucketName, prefix, startAfter, continuationToken)
	*objList = append(*objList, result.Contents...)
	if *result.IsTruncated {
		s3.GetObjectsListFromS3(bucketName, prefix, startAfter, *result.NextContinuationToken, objList)
	}
}

func (s *S3) DownloadFileFromS3(bucketName, key string) []byte {
	//https://stackoverflow.com/questions/46019484/buffer-implementing-io-writerat-in-go
	buf := aws.NewWriteAtBuffer([]byte{})
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	downloader := s3manager.NewDownloader(s.getSession())

	numBytes, err := downloader.Download(buf, input)
	fmt.Println("downloaded numBytes:", numBytes)
	if err != nil {
		logging.Error(err.Error())
	}
	return buf.Bytes()
}

func (s *S3) GenerateThumbsInS3(bucketName, prefix string, width, height int) {
	// list objects，获得全部的objects
	var objList []*s3.Object
	s.GetObjectsListFromS3(bucketName, prefix, "", "", &objList)
	fmt.Println("总计找到多少个文件(含目录数量): ", len(objList))
	// 遍历list，将list转为map，以方面下面的查找
	maps := make(map[string]*s3.Object)
	for _, obj := range objList {
		// 如果不是目录名，是文件名
		if *obj.Size != int64(0) {
			maps[*obj.Key] = obj
		}
	}
	// check list中是否有他的缩略图，如果已经是缩略图，则跳过
	for _, obj := range objList {
		// 如果不是目录名，是文件名
		if *obj.Size != int64(0) {
			// 先判断本身是不是缩略图，即本身是不是带_thumb的
			objName := *obj.Key
			// 如果本身不是缩略图，去map中找是不是有_thumb（即是不是有缩略图）
			if !image.IsThumbFileName(objName) {
				thumbName := image.GetThumbnailName(objName) // 算出如果有thumbnail，应该是什么文件名
				// 去找这个thumbnail文件名存在不存在
				if _, ok := maps[thumbName]; !ok {
					fmt.Println("缩略图" + thumbName + "不存在，创建缩略图...")
					// 如果没有缩略图，则下载原图，resize，上传缩略图（需注意缩略图不同的尺寸）
					originalByte := s.DownloadFileFromS3(bucketName, objName) // 下载原图
					// 生成缩略图
					fileBaseName := filepath.Base(objName)
					path, thumbBaseName := image.GenerateThumbBaseName(objName)
					// 生成缩略图
					buff, err := image.ResizeImage(
						bytes.NewReader(originalByte), fileBaseName, width, height, false, false)
					if err != nil {
						logging.Error(err.Error())
					}

					// 上传缩略图
					_, _ = s.AddFileToS3(&buff, thumbBaseName, path)

				} else {
					fmt.Println("缩略图" + thumbName + "存在！")
				}

			} else {
				//	如果本身是缩略图文件，去map中找是不是存在原文件，如果不存在，说明需要删掉这个缩略图。
				originName := image.GetOriginalName(objName)
				// 去找这个原文件名存在不存在
				if _, ok := maps[originName]; !ok {
					// 如果不存在，说明需要删掉这个缩略图。
					fmt.Println("原图" + originName + "不存在，删除缩略图...")
					s.DeleteFileFromS3(s.URL + objName)
				}

			}
		}
	}
}

// https://docs.aws.amazon.com/zh_cn/sdk-for-go/v1/developer-guide/s3-example-presigned-urls.html
// https://medium.com/@aidan.hallett/securing-aws-s3-uploads-using-presigned-urls-aa821c13ae8d
// Generate a Pre-Signed URL for an Amazon S3 PUT Operation with a Specific Payload
// You can generate a pre-signed URL for a PUT operation that checks whether users upload the
// correct content. When the SDK pre-signs a request, it computes the checksum of the request
// body and generates an MD5 checksum that is included in the pre-signed URL. Users must upload
// the same content that produces the same MD5 checksum generated by the SDK; otherwise, the
// operation fails. This is not the Content-MD5, but the signature. To enforce Content-MD5,
// simply add the header to the request.
// fileName 是文件名
// backetDir 是上传到bucket里的哪个文件夹，例如 "abc/upload"
// ! 注意：无法在此时将ACL设为"public-read"，需要通过cloudfront来处理
func (s *S3) GetPresigndURLForUpload(fileName, bucketDir string) (string, error) {
	svc := s3.New(s.getSession())
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(bucketDir + "/" + fileName),
		// Body:   strings.NewReader("EXPECTED CONTENTS"),
	})

	str, err := req.Presign(15 * time.Minute)

	// log.Println("The URL is:", str, " err:", err)

	return str, err
}

func (s *S3) GetPresignedURLForDownload(fileName, bucketDir string) (string, error) {
	svc := s3.New(s.getSession())

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(bucketDir + "/" + fileName),
	})
	str, err := req.Presign(15 * time.Minute)

	// log.Println("The URL is:", str, " err:", err)

	return str, err
}
