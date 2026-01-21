package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	addr       = ":8080"
	s3Endpoint = "http://localhost:9000"
	bucketName = "tweet-images"
)

func main() {
	globalCtx := context.Background()
	credentialsProvider := credentials.NewStaticCredentialsProvider("minio", "minio123", "")
	s3Config, err := config.LoadDefaultConfig(
		globalCtx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentialsProvider),
	)
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(s3Config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3Endpoint)
		o.UsePathStyle = true
		o.Credentials = aws.NewCredentialsCache(credentialsProvider)
	})
	presigner := s3.NewPresignClient(s3Client)

	if err := ensureBucket(globalCtx, s3Client, bucketName); err != nil {
		log.Fatal(err)
	}

	tweetRepo := NewTweetRepo()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /tweets:prepare-image-upload", func(w http.ResponseWriter, r *http.Request) {
		imageID := uuid.NewString()
		post, err := presigner.PresignPostObject(r.Context(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    &imageID,
		}, func(ppo *s3.PresignPostOptions) {
			ppo.Expires = 15 * time.Minute
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(PrepareTweetImageUploadResponse{
			ImageID: imageID,
			URL:     post.URL,
			Fields:  post.Values,
		})
	})
	mux.HandleFunc("POST /tweets", func(w http.ResponseWriter, r *http.Request) {
		var req CreateTweetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tweetRepo.Create(req.Content, req.ImageID)
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("GET /tweets", func(w http.ResponseWriter, r *http.Request) {
		tweets := tweetRepo.List()

		var resTweets []Tweet
		for _, tweet := range tweets {
			resTweets = append(resTweets, Tweet{
				ID:        tweet.ID,
				Content:   tweet.Content,
				ImageURL:  generateImageURL(tweet.ImageID),
				CreatedAt: tweet.CreatedAt,
			})
		}

		json.NewEncoder(w).Encode(ListTweetsResponse{
			Items: resTweets,
		})
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	log.Println("Start server on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func generateImageURL(imageID string) string {
	if imageID == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", s3Endpoint, bucketName, imageID)
}

func ensureBucket(ctx context.Context, client *s3.Client, bucketName string) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &bucketName,
	})
	if err == nil {
		return nil
	}

	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &bucketName,
	})
	return err
}

type PrepareTweetImageUploadResponse struct {
	ImageID string            `json:"imageId"`
	URL     string            `json:"url"`
	Fields  map[string]string `json:"fields"`
}

type CreateTweetRequest struct {
	Content string `json:"content"`
	ImageID string `json:"imageId"`
}

type CreateTweetResponse struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"imageUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

type ListTweetsResponse struct {
	Items []Tweet `json:"items"`
}

type Tweet struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"imageUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

type TweetRepo struct {
	tweets []*TweetEntity
}

func NewTweetRepo() *TweetRepo {
	return &TweetRepo{}
}

func (tr *TweetRepo) Create(content string, imageID string) {
	tr.tweets = append(tr.tweets, &TweetEntity{
		ID:        uuid.New().String(),
		Content:   content,
		ImageID:   imageID,
		CreatedAt: time.Now(),
	})
}

func (tr *TweetRepo) List() []*TweetEntity {
	return tr.tweets
}

type TweetEntity struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	ImageID   string    `json:"imageId"`
	CreatedAt time.Time `json:"createdAt"`
}
