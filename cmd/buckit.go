package cmd

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"gocloud.dev/blob/s3blob"
	"log"
	"net/http"
	"time"

	_ "gocloud.dev/blob/s3blob"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type buckit struct {
	HostName        string // maps to the dns hostname
	BucketName      string // s3 bucket name
	AccessKeyID     string // optional - if not given, will use env. If given, will use only for the bucket
	SecretAccessKey string // optional - if not given, will use env. If given, will use only for the bucket
	Region          string // required bucket region
}

type config struct {
	Buckits         []buckit
	Port            int
	ShutdownTimeout time.Duration
}

type App struct {
	Config config
	server *http.Server
}

func (a *App) Start() {
	a.server = &http.Server{
		Addr:           fmt.Sprintf(":%d", a.Config.Port),
		Handler:        a.handlerFunc(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Ready to listen on: %s", a.server.Addr)
	_ = a.server.ListenAndServe()
}

func (a *App) Stop() error {
	ctx, _ := context.WithTimeout(context.Background(), a.Config.ShutdownTimeout)
	return a.server.Shutdown(ctx)
}

func (a *App) handlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not supported", http.StatusMethodNotAllowed)
			return
		}
		for _, b := range a.Config.Buckits {
			if b.HostName == r.Host {
				fetchResource(w, r, b)
				return
			}
		}
		http.Error(w, "resource not found", http.StatusNotFound)
	}
}

func fetchResource(w http.ResponseWriter, r *http.Request, b buckit) {
	path := r.URL.Path
	if path == "" {
		path = "index.html"
	}
	// Establish an AWS session.
	// See https://docs.aws.amazon.com/sdk-for-go/api/aws/session/ for more info.
	// The region must match the region for "my-bucket".
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
		Region:      aws.String(b.Region),
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "cannot connect to bucket", http.StatusInternalServerError)
		return
	}

	bucket, err := s3blob.OpenBucket(r.Context(), sess, b.BucketName, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "cannot connect to bucket", http.StatusInternalServerError)
		return
	}

	br, err := bucket.NewReader(r.Context(), path, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "cannot find data", http.StatusNotFound)
		return
	}
	defer br.Close()

	_, err = br.WriteTo(w)
	if err != nil {
		log.Println(err)
		http.Error(w, "cannot fetch data", http.StatusInternalServerError)
		return
	}
	defer bucket.Close()
}
