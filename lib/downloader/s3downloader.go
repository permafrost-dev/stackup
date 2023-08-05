package downloader

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/minio/minio-go/pkg/s3utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Url struct {
	Endpoint   string
	BucketName string
	FileName   string
}

func ParseS3Url(urlstr string) (*S3Url, error) {
	temp := strings.Replace(urlstr, "s3://", "", 1)
	temp = strings.Replace(temp, "s3:", "", 1)
	temp = "s3://" + temp

	parsedUrl, err := url.Parse(urlstr)

	if err != nil {
		return nil, err
	}

	// Extract the endpoint, bucket name, and filename from the URL
	endpoint := parsedUrl.Host
	bucketName := parsedUrl.Path
	fileName := ""

	// Remove the leading slash from the bucket name
	if strings.HasPrefix(bucketName, "/") {
		bucketName = bucketName[1:]
	}

	// Extract the filename from the bucket name
	if strings.Contains(bucketName, "/") {
		parts := strings.SplitN(bucketName, "/", 2)
		bucketName = parts[0]
		fileName = parts[1]
	}

	// Create a new S3Url struct
	s3UrlStruct := &S3Url{
		Endpoint:   endpoint,
		BucketName: bucketName,
		FileName:   fileName,
	}

	return s3UrlStruct, nil
}

func ReadS3FileContents(s3url string, accessKey string, secretKey string, secure bool) string {
	data, err := ParseS3Url(s3url)

	s3Client, err := minio.New(data.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})

	if err != nil {
		fmt.Printf("%v", err)
		return ""
	}

	opts := minio.GetObjectOptions{}
	content, _ := FGetObject(*s3Client, context.Background(), data.BucketName, data.FileName, opts)

	return content
}

func FGetObject(client minio.Client, ctx context.Context, bucketName string, objectName string, opts minio.GetObjectOptions) (string, error) {
	// Input validation.
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return "", err
	}

	opts.SetRange(0, 8192*32)

	objectReader, err := client.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return "", err
	}

	objectReader.Seek(0, 0)
	content := make([]byte, 8192*32)
	objectReader.Read(content)

	return removeControlCharacters(string(content)), nil
}

func removeControlCharacters(input string) string {
	controlCharRegexp := regexp.MustCompile(`[\x00-\x08\x1A]`)
	return controlCharRegexp.ReplaceAllString(input, "")
}
