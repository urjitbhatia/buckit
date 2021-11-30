# Buckit

Buckit is a simple "server" that sits in front of S3 and serves protected content - no need to expose your S3 bucket to
the world or enable bucket website mode.

## Features

- Multiple buckets are supported
- Each bucket can have custom credentials
- Dynamic config

## Running Buckit

1. Create a buckit config file: `.buckit.yaml`. Example:

```yaml
Port: 9006
ShutdownTimeout: 20s
Buckits:
  - HostName: foo.com
    BucketName: foo-source-bucket
    Region: us-east-1
  - HostName: bar.com
    BucketName: bar-source-bucket
    Region: us-east-1
```

2. Run buckit: `go run main.go` 
