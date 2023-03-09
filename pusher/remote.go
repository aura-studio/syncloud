package pusher

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Remote interface {
	Push([]Pair)
}

type S3Remote struct {
	bucket string
}

func NewS3Remote(bucket string) *S3Remote {
	return &S3Remote{
		bucket: bucket,
	}
}

func (r *S3Remote) createS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create client, %v", err)
	}

	return s3.NewFromConfig(cfg), nil
}

func (r *S3Remote) uploadFileToS3(remoteFilePath string, localFilePath string) error {
	client, err := r.createS3Client()
	if err != nil {
		log.Panicf("failed to create s3 client, %v", err)
	}

	// Open local file for use.
	f, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", localFilePath, err)
	}
	defer f.Close()

	// Upload file body to S3.
	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(remoteFilePath),
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	return nil
}

func (r *S3Remote) batchUploadFilesToS3(pairs []Pair) error {
	var wg sync.WaitGroup
	var errChan = make(chan error, len(pairs))
	for _, pair := range pairs {
		wg.Add(1)
		go func(pair Pair) {
			defer wg.Done()
			localFilePath := pair.LocalFilePath
			remoteFilePath := pair.RemoteFilePath

			if stat, err := os.Stat(localFilePath); err != nil {
				if os.IsNotExist(err) {
					log.Printf("%s does not exist", localFilePath)
					errChan <- err
				} else {
					log.Printf("failed to stat file, %v", err)
					errChan <- err
					return
				}
			} else if stat.Size() == 0 {
				log.Printf("%s is empty", localFilePath)
				errChan <- err
				return
			} else {
				log.Printf("%s exists, uploading to s3[%s]...", localFilePath, remoteFilePath)
				if err := r.uploadFileToS3(remoteFilePath, localFilePath); err != nil {
					log.Printf("failed to upload file to s3, %v", err)
					errChan <- err
					return
				}
			}
		}(pair)
	}
	wg.Wait()

	if len(errChan) > 0 {
		return fmt.Errorf("%d errors occurred during uploading", len(errChan))
	}

	return nil
}

func (r *S3Remote) Push(tasks []Pair) {
	if err := r.batchUploadFilesToS3(tasks); err != nil {
		log.Panicf("failed to upload files to s3, %v", err)
	}
}
