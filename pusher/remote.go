package pusher

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Remote interface {
	Push(RemoteTask)
}

type RemoteTask struct {
	Pairs       []Pair
	Concurrency int
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
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRetryMode(aws.RetryModeAdaptive),
		config.WithRetryMaxAttempts(100000000),
		config.WithHTTPClient(&http.Client{
			Timeout: 86400 * time.Second,
		}),
	)
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

	contentType := strings.Split(mime.TypeByExtension(filepath.Ext(localFilePath)), ";")[0]

	log.Printf("%s exists, uploading to s3[%s(%s)]...", localFilePath, remoteFilePath, contentType)

	// Upload file body to S3.
	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(remoteFilePath),
		Body:        f,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	return nil
}

func (r *S3Remote) batchUploadFilesToS3(remoteTask RemoteTask) error {
	var errChan = make(chan error, len(remoteTask.Pairs))
	var taskCh = make(chan Pair, len(remoteTask.Pairs))
	var doneCh = make(chan struct{}, len(remoteTask.Pairs))
	for _, pair := range remoteTask.Pairs {
		taskCh <- pair
	}

	for i := 0; i < remoteTask.Concurrency; i++ {
		go func() {
			for pair := range taskCh {
				localFilePath := pair.LocalFilePath
				remoteFilePath := pair.RemoteFilePath

				if stat, err := os.Stat(localFilePath); err != nil {
					if os.IsNotExist(err) {
						log.Printf("%s does not exist", localFilePath)
						errChan <- err
						doneCh <- struct{}{}
					} else {
						log.Printf("failed to stat file, %v", err)
						errChan <- err
						doneCh <- struct{}{}
						return
					}
				} else if stat.Size() == 0 {
					log.Printf("%s is empty", localFilePath)
					errChan <- err
					doneCh <- struct{}{}
					return
				} else {
					if err := r.uploadFileToS3(remoteFilePath, localFilePath); err != nil {
						log.Printf("failed to upload file to s3, %v", err)
						errChan <- err
						doneCh <- struct{}{}
						return
					}
					doneCh <- struct{}{}
				}
			}
		}()
	}
	for len(doneCh) < len(remoteTask.Pairs) {
		time.Sleep(time.Millisecond)
	}

	if len(errChan) > 0 {
		return fmt.Errorf("%d errors occurred during uploading", len(errChan))
	}

	return nil
}

func (r *S3Remote) Push(remoteTask RemoteTask) {
	if err := r.batchUploadFilesToS3(remoteTask); err != nil {
		log.Panicf("failed to upload files to s3, %v", err)
	}
}
