// Bucket-level operations

package pkg

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const minMultipartPartSize int64 = 1024 * 1024 * 5

// R2Bucket represents a Cloudflare R2 bucket, storing the bucket's name and R2 client used to
// access the bucket.
type R2Bucket struct {
	Client *R2Client
	Name   string
}

// Bucket receives an R2 client and takes a bucket name as an argument, returning a configured
// R2Bucket struct. This allows for simple bucket-level operations. For example, you can create
// a bucket struct as so:
//
//	client := Client(Config{...})
//	bucket := client.Bucket("my-bucket")
//
//	// Then, you can perform bucket-level operations easily
//	bucket.Put("my-local-file.txt", "my-remote-file.txt")
func (c *R2Client) Bucket(bucketName string) R2Bucket {
	return R2Bucket{
		Client: c,
		Name:   bucketName,
	}
}

// GetObjects returns a list of all objects in a bucket. This method leverages S3's ListObjectsV2
// API call. The returned list of objects is of type types.Object, which is a struct containing all
// available information about the object, such as its name, size, and last modified date.
// This function properly handles pagination to retrieve all objects, even if there are more than 1000.
func (b *R2Bucket) GetObjects() []types.Object {
	return b.GetObjectsWithPrefix("")
}

// GetObjectsWithPrefix returns a list of objects in a bucket that have the specified prefix.
// This method leverages S3's ListObjectsV2 API call with the Prefix parameter to filter results.
// The returned list of objects is of type types.Object, which is a struct containing all
// available information about the object, such as its name, size, and last modified date.
// This function properly handles pagination to retrieve all objects, even if there are more than 1000.
func (b *R2Bucket) GetObjectsWithPrefix(prefix string) []types.Object {
	var allObjects []types.Object
	var continuationToken *string

	for {
		input := &s3.ListObjectsV2Input{
			Bucket: &b.Name,
		}

		// Add prefix if provided
		if prefix != "" {
			input.Prefix = &prefix
		}

		// Add continuation token if we have one from previous iteration
		if continuationToken != nil {
			input.ContinuationToken = continuationToken
		}

		listObjectsOutput, err := b.Client.ListObjectsV2(context.TODO(), input)
		if err != nil {
			log.Fatal(err)
		}

		// Append the objects from this page to our complete list
		allObjects = append(allObjects, listObjectsOutput.Contents...)

		// Check if there are more pages to fetch
		if listObjectsOutput.IsTruncated != nil && *listObjectsOutput.IsTruncated {
			continuationToken = listObjectsOutput.NextContinuationToken
		} else {
			// No more pages, we're done
			break
		}
	}

	return allObjects
}

// GetObjectPaths returns a list of all object paths in a bucket, represented as strings. This
// method is a wrapper around GetObjects, which returns a list of types.Object structs.
func (b *R2Bucket) GetObjectPaths() []string {
	var objectPaths []string
	for _, object := range b.GetObjects() {
		objectPaths = append(objectPaths, *object.Key)
	}
	return objectPaths
}

// PrintObjects prints a list of all objects in a bucket. This method is a wrapper around GetObjects,
// which returns a list of types.Object structs. The returned list of objects is formatted as a table
// with the following columns: last modified date, file size, file name. The file size column is
// formatted as a string with the file size and its unit (e.g. 1.2 MB).
func (b *R2Bucket) PrintObjects() {
	// Get creation date, file size, and name of each object
	var objectData [][]string
	for _, object := range b.GetObjects() {
		// Get file size
		fs := fileSizeFmt(*object.Size)

		// Append last modified, file size, and file name to objectData
		objectData = append(objectData, []string{
			object.LastModified.Format("2006-01-02 15:04:05"),
			fs[0],
			fs[1],
			*object.Key,
		})
	}

	// Get length of longest file size string
	var longestFileSizeString int
	var longestFileSizeUnitString int
	for _, object := range objectData {
		if len(object[1]) > longestFileSizeString {
			longestFileSizeString = len(object[1])
		}
		if len(object[2]) > longestFileSizeUnitString {
			longestFileSizeUnitString = len(object[2])
		}
	}

	// Print objects
	for _, object := range objectData {
		fmt.Println(
			object[0],
			strings.Repeat(" ", longestFileSizeString-len(object[1])),
			object[1],
			object[2],
			strings.Repeat(" ", longestFileSizeUnitString-len(object[2])),
			object[3],
		)
	}
}

// Put puts an object into a bucket. The inputted object is represented as an io.Reader, which can
// be created from a file, a string, or any other type that implements the io.Reader interface. The
// bucketPath argument takes the path for the object to be put in the bucket.
func (b *R2Bucket) Put(file io.Reader, bucketPath string) error {
	_, err := b.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(b.Name),
		Key:    aws.String(bucketPath),
		Body:   file,
	})
	return err
}

// Upload uploads a local file to a bucket. The localPath argument takes the path to the local file
// to be uploaded. The bucketPath argument takes the path for the object to be put in the bucket.
// This method is a wrapper around Put, which takes an io.Reader as an argument.
func (b *R2Bucket) Upload(localPath, bucketPath string) {
	file, err := os.Open(localPath)
	if err != nil {
		log.Fatalf("Couldn't open file %s to upload: %v\n", localPath, err)
	}

	defer file.Close()

	err = b.Put(file, bucketPath)
	if err != nil {
		log.Fatalf("Couldn't upload file %s to r2://%s/%s: %v\n", localPath, b.Name, bucketPath, err)
	}
}

// PutStream uploads a stream to a bucket using multipart upload for efficient streaming.
// This method is optimized for streaming data from sources like stdin where the size is unknown.
// The partSize parameter controls the size of each part in bytes (minimum 5MB).
// The concurrency parameter controls how many parts are uploaded in parallel.
func (b *R2Bucket) PutStream(reader io.Reader, bucketPath string, partSize int64, concurrency int) error {
	if partSize > 0 && partSize < minMultipartPartSize {
		return fmt.Errorf("part size must be at least %d bytes", minMultipartPartSize)
	}

	uploader := transfermanager.New(&b.Client.Client, func(o *transfermanager.Options) {
		if partSize > 0 {
			o.PartSizeBytes = partSize
			o.MultipartUploadThreshold = partSize
		}
		if concurrency > 0 {
			o.Concurrency = concurrency
		}
	})

	_, err := uploader.UploadObject(context.TODO(), &transfermanager.UploadObjectInput{
		Bucket: aws.String(b.Name),
		Key:    aws.String(bucketPath),
		Body:   reader,
	})

	return err
}

// Get gets an object from a bucket. The bucketPath argument takes the path to the object in the
// bucket. This method returns an io.ReadCloser, which can be used to read the object's contents.
// This method is a wrapper around the S3 GetObject API call.
func (b *R2Bucket) Get(bucketPath string) io.ReadCloser {
	obj, err := b.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(b.Name),
		Key:    aws.String(bucketPath),
	})
	if err != nil {
		log.Fatalf("Couldn't get file r2://%s/%s: %v\n", b.Name, bucketPath, err)
	}

	return obj.Body
}

// Download downloads an object from a bucket to a local file. The bucketPath argument takes the
// path to the object in the bucket. The localPath argument takes the path to the local file to
// download to. This method is a wrapper around Get, which returns an io.ReadCloser.
func (b *R2Bucket) Download(bucketPath, localPath string) {
	objBody := b.Get(bucketPath)

	file, err := os.Create(localPath)
	if err != nil {
		log.Fatalf("Couldn't create file %s to download to: %v\n", localPath, err)
	}

	defer file.Close()
	_, err = io.Copy(file, objBody)
	if err != nil {
		log.Fatalf("Couldn't download file r2://%s/%s to %s: %v\n", b.Name, bucketPath, localPath, err)
	}
}

// Copy copies an object from a bucket to another bucket. The bucketPath argument takes the path to
// the object in the bucket. The copyToURI argument takes the URI of the bucket to copy the object
// to. This method is a wrapper around the S3 CopyObject API call.
func (b *R2Bucket) Copy(bucketPath string, copyToURI R2URI) {
	_, err := b.Client.CopyObject(context.TODO(), &s3.CopyObjectInput{
		Bucket:     aws.String(copyToURI.Bucket),
		CopySource: aws.String(b.Name + "/" + bucketPath),
		Key:        aws.String(copyToURI.Path),
	})
	if err != nil {
		log.Fatalf("Couldn't copy file r2://%s/%s to r2://%s/%s: %v\n", b.Name, bucketPath, copyToURI.Bucket, copyToURI.Path, err)
	}
}

// Delete deletes an object from a bucket. The bucketPath argument takes the path to the object in
// the bucket. This method is a wrapper around the S3 DeleteObject API call.
func (b *R2Bucket) Delete(bucketPath string) {
	_, err := b.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(b.Name),
		Key:    aws.String(bucketPath),
	})
	if err != nil {
		log.Fatalf("Couldn't delete file r2://%s/%s: %v\n", b.Name, bucketPath, err)
	}
}

// SyncLocalToR2 syncs a local directory to an R2 bucket. The sourcePath argument takes the path to
// the local directory to sync. This method iterates through the local directory and uploads any new
// or changed files to the bucket.
func (b *R2Bucket) SyncLocalToR2(sourcePath string) {
	b.SyncLocalToR2WithPrefix(sourcePath, "")
}

// SyncLocalToR2WithPrefix syncs a local directory to an R2 bucket with a specific prefix.
// The sourcePath argument takes the path to the local directory to sync.
// The prefix argument specifies the prefix to add to all uploaded objects.
func (b *R2Bucket) SyncLocalToR2WithPrefix(sourcePath string, prefix string, filters ...func(string) bool) {
	// Check if source path exists and is a directory
	if !isDir(sourcePath) {
		log.Fatal("Source path must be a directory.")
	}

	// Ensure prefix ends with / if it's not empty
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	filter := firstFilter(filters...)

	// Get extant paths and their MD5 checksums in bucket with the specified prefix
	bucketObjects := make(map[string]string)
	for _, object := range b.GetObjectsWithPrefix(prefix) {
		bucketObjects[*object.Key] = strings.Trim(*object.ETag, `"`)
	}

	// Iterate through paths in source directory
	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If path is a file, upload it
		if !info.IsDir() {
			// Get relative path from source directory
			relativePath := strings.TrimPrefix(path, sourcePath)
			relativePath = strings.TrimPrefix(relativePath, "/")

			if filter != nil && !filter(relativePath) {
				return nil
			}

			// Add prefix to create final bucket path
			bucketPath := prefix + relativePath

			objectMD5, objectInBucket := bucketObjects[bucketPath]
			if !objectInBucket || (md5sum(path) != objectMD5) {
				b.Upload(path, bucketPath)
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

// SyncR2ToLocal syncs an R2 bucket to a local directory. The destinationPath argument takes the
// path to the local directory to sync. This method iterates through the bucket and downloads any
// new or changed files to the local directory.
func (b *R2Bucket) SyncR2ToLocal(destinationPath string) {
	b.SyncR2ToLocalWithPrefix(destinationPath, "")
}

// SyncR2ToLocalWithPrefix syncs objects from an R2 bucket with a specific prefix to a local directory.
// The destinationPath argument takes the path to the local directory to sync.
// The prefix argument specifies which objects to sync (only objects with this prefix).
func (b *R2Bucket) SyncR2ToLocalWithPrefix(destinationPath string, prefix string, filters ...func(string) bool) {
	// Check if destination path exists and is a directory
	if !isDir(destinationPath) {
		log.Fatal("Destination path must be a directory.")
	}

	filter := firstFilter(filters...)

	// Iterate through objects with the specified prefix and download necessary ones
	for _, object := range b.GetObjectsWithPrefix(prefix) {
		objectPath := *object.Key
		hash := strings.Trim(*object.ETag, `"`)

		// Remove prefix from object path to get relative path
		relativePath := objectPath
		if prefix != "" {
			relativePath = strings.TrimPrefix(objectPath, prefix)
			// Also remove leading slash if present
			relativePath = strings.TrimPrefix(relativePath, "/")
		}

		if filter != nil && !filter(relativePath) {
			continue
		}

		// Construct local file path
		localPath := filepath.Join(destinationPath, relativePath)

		// Security check: ensure the path is within the destination directory
		absLocalPath, err := filepath.Abs(localPath)
		if err != nil {
			log.Printf("Warning: could not resolve path %s: %v", localPath, err)
			continue
		}
		absDestPath, err := filepath.Abs(destinationPath)
		if err != nil {
			log.Printf("Warning: could not resolve destination path %s: %v", destinationPath, err)
			continue
		}
		if !strings.HasPrefix(absLocalPath, absDestPath+string(filepath.Separator)) && absLocalPath != absDestPath {
			log.Printf("Warning: skipping file %s - path traversal detected", objectPath)
			continue
		}

		// Check if file needs to be downloaded
		if !fileExists(localPath) || (fileExists(localPath) && (md5sum(localPath) != hash)) {
			ensureDirExists(localPath)
			b.Download(objectPath, localPath)
		}
	}
}

// SyncR2ToR2 syncs an R2 bucket to another R2 bucket. The destBucket argument takes the bucket to
// sync to. This method iterates through the bucket and copies any new or changed files to the
// destination bucket.
func (b *R2Bucket) SyncR2ToR2(destBucket R2Bucket) {
	b.SyncR2ToR2WithPrefix(destBucket, "", "")
}

// SyncR2ToR2WithPrefix syncs objects from an R2 bucket with a specific prefix to another R2 bucket.
// The sourcePrefix specifies which objects to sync from the source bucket.
// The destPrefix specifies the prefix to add to objects in the destination bucket.
func (b *R2Bucket) SyncR2ToR2WithPrefix(destBucket R2Bucket, sourcePrefix string, destPrefix string, filters ...func(string) bool) {
	// Ensure prefixes end with / if they're not empty
	if sourcePrefix != "" && !strings.HasSuffix(sourcePrefix, "/") {
		sourcePrefix = sourcePrefix + "/"
	}
	if destPrefix != "" && !strings.HasSuffix(destPrefix, "/") {
		destPrefix = destPrefix + "/"
	}

	filter := firstFilter(filters...)

	// Get extant paths and their MD5 checksums in source bucket with prefix
	sourceBucketObjects := make(map[string]string)
	for _, object := range b.GetObjectsWithPrefix(sourcePrefix) {
		sourceBucketObjects[*object.Key] = strings.Trim(*object.ETag, `"`)
	}

	// Get extant paths and their MD5 checksums in destination bucket with prefix
	destBucketObjects := make(map[string]string)
	for _, object := range destBucket.GetObjectsWithPrefix(destPrefix) {
		destBucketObjects[*object.Key] = strings.Trim(*object.ETag, `"`)
	}

	// Iterate through paths in source bucket and copy necessary ones
	for sourcePath, sourceHash := range sourceBucketObjects {
		// Calculate destination path
		relativePath := sourcePath
		if sourcePrefix != "" {
			relativePath = strings.TrimPrefix(sourcePath, sourcePrefix)
		}

		if filter != nil && !filter(relativePath) {
			continue
		}

		destPath := destPrefix + relativePath

		destHash, sourceObjectInDestBucket := destBucketObjects[destPath]
		if !sourceObjectInDestBucket || (sourceHash != destHash) {
			b.Copy(sourcePath, R2URI{Bucket: destBucket.Name, Path: destPath})
		}
	}
}

func firstFilter(filters ...func(string) bool) func(string) bool {
	for _, filter := range filters {
		if filter != nil {
			return filter
		}
	}
	return nil
}

// GetURL returns a presigned URL for an object to get from a bucket. The uri argument takes the
// URI of the object in the bucket. This method is a wrapper around the S3 PresignGetObject API
// call.
func (pc *R2PresignClient) GetURL(uri R2URI) string {
	presignResult, err := pc.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(uri.Bucket),
		Key:    aws.String(uri.Path),
	})
	if err != nil {
		log.Fatal("Couldn't get presigned URL for GetObject")
	}
	return presignResult.URL
}

// PutURL returns a presigned URL for an object to put in a bucket. The uri argument takes the URI
// of the object in the bucket. This method is a wrapper around the S3 PresignPutObject API call.
func (pc *R2PresignClient) PutURL(uri R2URI) string {
	presignResult, err := pc.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(uri.Bucket),
		Key:    aws.String(uri.Path),
	})
	if err != nil {
		log.Fatal("Couldn't get presigned URL for PutObject")
	}
	return presignResult.URL
}
