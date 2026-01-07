package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	addr = ":8080"
)

func main() {
	globalCtx := context.Background()
	s3Config, err := config.LoadDefaultConfig(globalCtx)
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(s3Config, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	presigner := s3.NewPresignClient(s3Client)

	tweetRepo := NewTweetRepo()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /tweets:prepare-image-upload", func(w http.ResponseWriter, r *http.Request) {
		post, err := presigner.PresignPostObject(r.Context(), &s3.PutObjectInput{}, func(ppo *s3.PresignPostOptions) {
			ppo.Expires = 15 * time.Minute
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(PrepareTweetImageUploadResponse{
			URL: post.URL,
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
			items: resTweets,
		})
	})

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func generateImageURL(imageID string) string {
	return fmt.Sprintf("https://example.com/images/%s", imageID)
}

type PrepareTweetImageUploadResponse struct {
	ImageID string `json:"imageId"`
	URL     string `json:"url"`
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
	items []Tweet
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
