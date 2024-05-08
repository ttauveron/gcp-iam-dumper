package gcp

import (
	"cloud.google.com/go/storage"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
)

// UploadFilesToGCS uploads the specified path to GCS. If the path is a directory, it uploads all files in the directory.
// If the path is a file, it uploads the single file.
func UploadFilesToGCS(ctx context.Context, bucketName, srcPath string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	info, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return uploadDirectory(ctx, bucket, srcPath)
	}
	return uploadFile(ctx, bucket, srcPath, filepath.Base(srcPath))
}

func uploadDirectory(ctx context.Context, bucket *storage.BucketHandle, srcDir string) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		srcPath := filepath.Join(srcDir, f.Name())
		if !f.IsDir() {
			if err := uploadFile(ctx, bucket, srcPath, f.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}

func uploadFile(ctx context.Context, bucket *storage.BucketHandle, srcPath, dstPath string) error {
	obj := bucket.Object(dstPath)
	w := obj.NewWriter(ctx)

	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = io.Copy(w, file); err != nil {
		w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	log.Printf("Uploaded %s to %s", srcPath, dstPath)
	return nil
}
