package storage

import (
	"bytes"
	"io"

	"github.com/HotCodeGroup/warscript-imageserver/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Storage struct {
	SVC        *s3.S3
	BucketName string
}

type FileInfo struct {
	FileUUID string
	Type     string
	Size     int64
}

// StorageInit inits file storage
func Init(awsAccess, awsSecret, awsToken, bucketName string) (*Storage, error) {
	creds := credentials.NewStaticCredentials(awsAccess, awsSecret, awsToken)
	_, err := creds.Get()
	if err != nil {
		return nil, errors.Wrap(err, "credentials value cannot be retrieved")
	}

	cfg := aws.NewConfig().WithRegion("eu-central-1").WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	storage := &Storage{
		SVC:        svc,
		BucketName: bucketName,
	}

	return storage, nil
}

func (s *Storage) SaveImage(file io.ReadSeeker, size int64) (string, error) {
	fileType, err := utils.GetImageType(file)
	if err != nil {
		return "", errors.Wrap(err, "detecting fyle type failed")
	}

	fileUUID := uuid.NewV4().String()

	buffer := make([]byte, size)
	if _, err = file.Read(buffer); err != nil {
		return "", errors.Wrap(err, "file read failed")
	}

	fileBytes := bytes.NewReader(buffer)
	_, err = s.SVC.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.BucketName),
		Key:           aws.String("/" + fileUUID),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	})
	if err != nil {
		return "", errors.Wrap(err, "load to s3 error")
	}

	return fileUUID, nil
}

func (s *Storage) GetFile(fileUUID string) (io.ReadCloser, *FileInfo, error) {
	resp, err := s.SVC.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String("/" + fileUUID),
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "can not get file from s3")
	}

	return resp.Body, &FileInfo{
		FileUUID: fileUUID,
		Type:     *resp.ContentType,
		Size:     *resp.ContentLength,
	}, nil
}
