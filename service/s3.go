package service

import (
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"github.com/zjyl1994/b2webp/common/vars"
)

var S3Service s3Service

type s3Service struct{}

func (s s3Service) getSession() (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(vars.S3Setting.AccessId, vars.S3Setting.AccessKey, ""),
		Endpoint:         aws.String(vars.S3Setting.Endpoint),
		Region:           aws.String(vars.S3Setting.Region),
		S3ForcePathStyle: aws.Bool(true),
	})
}

func (s s3Service) Get(remotePath, localPath string) error {
	remotePath = filepath.Join(vars.S3Setting.ObjectPrefix, remotePath)

	logrus.Infoln("S3Service GET", localPath, remotePath)

	sess, err := s.getSession()
	if err != nil {
		return err
	}

	downloader := s3manager.NewDownloader(sess)
	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(vars.S3Setting.Bucket),
			Key:    aws.String(remotePath),
		})
	return err
}

func (s s3Service) Put(localPath, remotePath, contentType, contentMD5 string) error {
	remotePath = filepath.Join(vars.S3Setting.ObjectPrefix, remotePath)

	logrus.Infoln("S3Service PUT", localPath, remotePath)

	fin, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer fin.Close()

	s3sess, err := s.getSession()
	if err != nil {
		return err
	}

	_, err = s3.New(s3sess).PutObject(&s3.PutObjectInput{
		Body:        fin,
		Bucket:      aws.String(vars.S3Setting.Bucket),
		Key:         aws.String(remotePath),
		ContentType: aws.String(contentType),
		ContentMD5:  aws.String(contentMD5),
	})

	return err
}

func (s s3Service) Delete(remotePath string) error {
	remotePath = filepath.Join(vars.S3Setting.ObjectPrefix, remotePath)

	logrus.Infoln("S3Service DELETE", remotePath)

	s3sess, err := s.getSession()
	if err != nil {
		return err
	}

	_, err = s3.New(s3sess).DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(vars.S3Setting.Bucket),
		Key:    aws.String(remotePath),
	})

	return err
}
