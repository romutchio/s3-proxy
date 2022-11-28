package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

func main() {
	AccessKey := os.Getenv("APP_S3_ACCESS_KEY_ID")
	SecretKey := os.Getenv("APP_S3_SECRET_ACCESS_KEY")
	sess, err := session.NewSession(&aws.Config{
		Endpoint:    aws.String("s3-proxy.query.consul"),
		Credentials: credentials.NewStaticCredentials(AccessKey, SecretKey, ""),
		Region:      aws.String(endpoints.UsWest2RegionID),
	})
	if err != nil {
		log.Fatal(err)
	}
	client := s3.New(sess)
	router := gin.Default()

	router.GET("/:bucket/*key", func(c *gin.Context) {
		bucket := c.Param("bucket")
		key := c.Param("key")
		resp, err := client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == s3.ErrCodeNoSuchKey {
				c.Status(http.StatusNotFound)
				return
			}
		}
		defer resp.Body.Close()
		if *resp.ETag == c.GetHeader("If-None-Match") {
			c.Status(http.StatusNotModified)
			return
		}
		extraHeaders := map[string]string{
			"etag":          *resp.ETag,
			"last-modified": (*resp.LastModified).Format(http.TimeFormat),
		}
		c.DataFromReader(http.StatusOK, *resp.ContentLength, *resp.ContentType, resp.Body, extraHeaders)
	})

	router.Run(":8000")
}