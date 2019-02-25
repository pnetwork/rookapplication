package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pivotal-golang/bytefmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	//"strconv"
	//rr "crypto/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Global variables
var access_key, secret_key, url_host, bucket, region, formKey string
var duration_secs, threads, loops, filesize, localloops, localloopsleep int
var object_size uint64
var object_data []byte
var object_data_md5 string
var running_threads, upload_count, download_count, delete_count, upload_slowdown_count, download_slowdown_count, delete_slowdown_count int32
var endtime, upload_finish, download_finish, delete_finish time.Time
var deleteobj bool
var wg sync.WaitGroup

func logit(msg string) {
	fmt.Println(msg)
	logfile, _ := os.OpenFile("steps-benchmark.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if logfile != nil {
		logfile.WriteString(time.Now().Format(http.TimeFormat) + ": " + msg + "\n")
		logfile.Close()
	}
}
func logresult(msg string) {
	fmt.Println(msg)
	//t := time.Now()
	//output := fmt.Sprintf("benchmark-result.log", "benchmark", bucket, t.Format("20060102150405"))
	//output := fmt.Sprintf("%s-%s-%s.log", "benchmark", bucket, t.Format("20060102150405"))
	logfile, _ := os.OpenFile("benchmark-result.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if logfile != nil {
		logfile.WriteString(time.Now().Format("20060102150405") + ": " + msg + "\n")
		logfile.Close()
	}
}

// Our HTTP transport used for the roundtripper below
var HTTPTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 0,
	// Allow an unlimited number of idle connections
	MaxIdleConnsPerHost: 4096,
	MaxIdleConns:        0,
	// But limit their idle time
	IdleConnTimeout: time.Minute,
	// Ignore TLS errors
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var httpClient = &http.Client{Transport: HTTPTransport}

func getS3Client() (*s3.S3, *session.Session) {
	// Build our config
	creds := credentials.NewStaticCredentials(access_key, secret_key, "")
	loglevel := aws.LogOff
	// Build the rest of the configuration
	awsConfig := &aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(url_host),
		Credentials:      creds,
		DisableSSL:       aws.Bool(true),
		LogLevel:         &loglevel,
		S3ForcePathStyle: aws.Bool(true),
		//S3Disable100Continue: aws.Bool(true),
		// Comment following to use default transport
		HTTPClient: &http.Client{Transport: HTTPTransport},
	}
	session := session.New(awsConfig)
	client := s3.New(session)
	if client == nil {
		log.Fatalf("FATAL: Unable to create new client.")
	}
	// Return success
	return client, session
}

func createBucket(ignore_errors bool) {
	// Get a client
	client, _ := getS3Client()
	// Create our bucket (may already exist without error)
	cbparams := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(""), //must to setup to "", cannot be region name just ""
			//LocationConstraint: aws.String("US"),
		},
	}
	//in := &s3.CreateBucketInput{Bucket: aws.String(bucket)}
	if _, err := client.CreateBucket(cbparams); err != nil {
		if ignore_errors {
			log.Printf("WARNING: createBucket %s error, ignoring %v", bucket, err)
		} else {
			log.Fatalf("FATAL: Unable to create bucket %s (is your access and secret correct?): %v", bucket, err)
		}
	}
}

func deleteAllObjects() {
	// Get a client
	client, _ := getS3Client()
	// Use multiple routines to do the actual delete
	var doneDeletes sync.WaitGroup
	// Loop deleting our versions reading as big a list as we can
	var keyMarker, versionId *string
	var err error
	for loop := 1; ; loop++ {
		// Delete all the existing objects and versions in the bucket
		in := &s3.ListObjectVersionsInput{Bucket: aws.String(bucket), KeyMarker: keyMarker, VersionIdMarker: versionId, MaxKeys: aws.Int64(1000)}
		if listVersions, listErr := client.ListObjectVersions(in); listErr == nil {
			delete := &s3.Delete{Quiet: aws.Bool(true)}
			for _, version := range listVersions.Versions {
				delete.Objects = append(delete.Objects, &s3.ObjectIdentifier{Key: version.Key, VersionId: version.VersionId})
			}
			for _, marker := range listVersions.DeleteMarkers {
				delete.Objects = append(delete.Objects, &s3.ObjectIdentifier{Key: marker.Key, VersionId: marker.VersionId})
			}
			if len(delete.Objects) > 0 {
				// Start a delete routine
				doDelete := func(bucket string, delete *s3.Delete) {
					if _, e := client.DeleteObjects(&s3.DeleteObjectsInput{Bucket: aws.String(bucket), Delete: delete}); e != nil {
						err = fmt.Errorf("DeleteObjects unexpected failure: %s", e.Error())
					}
					doneDeletes.Done()
				}
				doneDeletes.Add(1)
				go doDelete(bucket, delete)
			}
			// Advance to next versions
			if listVersions.IsTruncated == nil || !*listVersions.IsTruncated {
				break
			}
			keyMarker = listVersions.NextKeyMarker
			versionId = listVersions.NextVersionIdMarker
		} else {
			// The bucket may not exist, just ignore in that case
			if strings.HasPrefix(listErr.Error(), "NoSuchBucket") {
				return
			}
			err = fmt.Errorf("ListObjectVersions unexpected failure: %v", listErr)
			break
		}
	}
	// Wait for deletes to finish
	doneDeletes.Wait()
	// If error, it is fatal
	if err != nil {
		log.Fatalf("FATAL: Unable to delete objects from bucket: %v", err)
	}
}

// canonicalAmzHeaders -- return the x-amz headers canonicalized
func canonicalAmzHeaders(req *http.Request) string {
	// Parse out all x-amz headers
	var headers []string
	for header := range req.Header {
		norm := strings.ToLower(strings.TrimSpace(header))
		if strings.HasPrefix(norm, "x-amz") {
			headers = append(headers, norm)
		}
	}
	// Put them in sorted order
	sort.Strings(headers)
	// Now add back the values
	for n, header := range headers {
		headers[n] = header + ":" + strings.Replace(req.Header.Get(header), "\n", " ", -1)
	}
	// Finally, put them back together
	if len(headers) > 0 {
		return strings.Join(headers, "\n") + "\n"
	} else {
		return ""
	}
}

func hmacSHA1(key []byte, content string) []byte {
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

func setSignature(req *http.Request) {
	// Setup default parameters
	dateHdr := time.Now().UTC().Format("20060102T150405Z")
	req.Header.Set("X-Amz-Date", dateHdr)
	// Get the canonical resource and header
	canonicalResource := req.URL.EscapedPath()
	canonicalHeaders := canonicalAmzHeaders(req)
	stringToSign := req.Method + "\n" + req.Header.Get("Content-MD5") + "\n" + req.Header.Get("Content-Type") + "\n\n" +
		canonicalHeaders + canonicalResource
	hash := hmacSHA1([]byte(secret_key), stringToSign)
	signature := base64.StdEncoding.EncodeToString(hash)
	req.Header.Set("Authorization", fmt.Sprintf("AWS %s:%s", access_key, signature))
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandStringBytes(n int) string {
	fmt.Println("prepare data now........")
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	fmt.Println("prepare data done.......")
	return string(b)
}

func deleteObjects(thread_num int, loop_num int, bucket string, formKey string) {
	svc, _ := getS3Client()
	//_, svc := getS3Client()
	//fmt.Println(b)
	tkey := fmt.Sprintf("%s-%d-%d", formKey, thread_num, loop_num)
	input := &s3.DeleteObjectInput{
		//	Body:   aws.ReadSeekCloser(strings.NewReader("filetoupload")),
		Bucket: aws.String(bucket),
		Key:    aws.String(tkey),
	}
	_, err := svc.DeleteObject(input)
	if err != nil {
		fmt.Println("some error")
		fmt.Println(err)
		return
	}
	fmt.Println("delete object", tkey)
}

func cleanAllObjects(loops int, threads int, bucket string, formKey string) {
	for loop := 0; loop < localloops; loop++ {
		for thread := 0; thread < threads; thread++ {
			deleteObjects(thread, loop, bucket, formKey)
		}
	}
}

func putObject(thread_num int, bucket string, b string, formKey string) {
	//svc := s3.New(session.New())
	logit("now in putobject thread")
	defer wg.Done()
	svc, _ := getS3Client()
	//_, svc := getS3Client()
	//formKey := "objkey"
	//fmt.Println(b)
	for localloop_num := 0; localloop_num < localloops; localloop_num++ {
		tkey := fmt.Sprintf("%s-%d-%d", formKey, thread_num, localloop_num)
		fmt.Println("threadnum:", thread_num, "uploadfile.....")
		input := &s3.PutObjectInput{
			//	Body:   aws.ReadSeekCloser(strings.NewReader("filetoupload")),
			Body:   aws.ReadSeekCloser(strings.NewReader(b)),
			Bucket: aws.String(bucket),
			Key:    aws.String(tkey),
		}
		_, err := svc.PutObject(input)
		if err != nil {
			fmt.Println("some error")
			fmt.Println(err)
			return
		}
		//fmt.Println(result)
		time.Sleep(time.Duration(localloopsleep) * time.Second)
	}
	atomic.AddInt32(&running_threads, -1)
}

func runUploadNew(thread_num int, loop_num int, bucket string) {
	atomic.AddInt32(&upload_count, 1)
	atomic.AddInt32(&running_threads, -1)
	s3Client, newSession := getS3Client()
	cparams := &s3.HeadBucketInput{
		Bucket: aws.String(bucket), // Required
	}
	_, err := s3Client.HeadBucket(cparams)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	uploader := s3manager.NewUploader(newSession)
	filename := "/tmp/demo.pdf" //上传文件路径
	f, err := os.Open(filename)
	if err != nil {
		fmt.Errorf("failed to open file %q, %v", filename, err)
		return
	}
	formKey := "go-mynewdata.log"
	tkey := fmt.Sprintf("%s-%d-%d", formKey, thread_num, loop_num)
	key := aws.String(tkey)
	myContentType := aws.String("application/zip") //content-type设置
	myACL := aws.String("public-read")             //acl 设置
	metadata_key := "udf-metadata"                 //自定义Metadata key
	metadata_value := "abc"                        //自定义Metadata value
	myMetadata := map[string]*string{
		metadata_key: &metadata_value,
	}
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         key,
		Body:        f,
		ContentType: myContentType,
		ACL:         myACL,
		Metadata:    myMetadata,
	}, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 分块大小,当文件体积超过10M开始进行分块上传
		u.LeavePartsOnError = true
		u.Concurrency = 3
	}) //并发数
	if err != nil {
		fmt.Printf("Failed to upload data to %s/%s, %s\n", *aws.String(bucket), *key, err.Error())
		return
	}
	atomic.AddInt32(&upload_count, -1)
	fmt.Printf("file uploaded to, %s\n", result.Location)

	atomic.AddInt32(&running_threads, -1)
}
func runUpload(thread_num int) {
	for time.Now().Before(endtime) {
		objnum := atomic.AddInt32(&upload_count, 1)
		fileobj := bytes.NewReader(object_data)
		prefix := fmt.Sprintf("%s/%s/Object-%d", url_host, bucket, objnum)
		req, _ := http.NewRequest("PUT", prefix, fileobj)
		fmt.Println("show object size %s", object_size)
		//req.Header.Set("Content-Length", strconv.FormatUint(object_size, 10))
		req.Header.Set("Content-Length", "1024") // strconv.FormatUint(object_size, 10))
		req.Header.Set("Content-MD5", object_data_md5)
		setSignature(req)
		if resp, err := httpClient.Do(req); err != nil {
			log.Fatalf("FATAL: Error uploading object %s: %v", prefix, err)
		} else if resp != nil && resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusServiceUnavailable {
				atomic.AddInt32(&upload_slowdown_count, 1)
				atomic.AddInt32(&upload_count, -1)
			} else {
				fmt.Printf("Upload status %s: resp: %+v\n", resp.Status, resp)
				if resp.Body != nil {
					body, _ := ioutil.ReadAll(resp.Body)
					fmt.Printf("Body: %s\n", string(body))
				}
			}
		}
	}
	// Remember last done time
	// One less thread
	atomic.AddInt32(&running_threads, -1)
}

func runDownload(thread_num int) {
	for time.Now().Before(endtime) {
		atomic.AddInt32(&download_count, 1)
		objnum := rand.Int31n(download_count) + 1
		prefix := fmt.Sprintf("%s/%s/Object-%d", url_host, bucket, objnum)
		req, _ := http.NewRequest("GET", prefix, nil)
		setSignature(req)
		if resp, err := httpClient.Do(req); err != nil {
			log.Fatalf("FATAL: Error downloading object %s: %v", prefix, err)
		} else if resp != nil && resp.Body != nil {
			if resp.StatusCode == http.StatusServiceUnavailable {
				atomic.AddInt32(&download_slowdown_count, 1)
				atomic.AddInt32(&download_count, -1)
			} else {
				io.Copy(ioutil.Discard, resp.Body)
			}
		}
	}
	// Remember last done time
	download_finish = time.Now()
	// One less thread
	atomic.AddInt32(&running_threads, -1)
}

func runDelete(thread_num int) {
	for {
		objnum := atomic.AddInt32(&delete_count, 1)
		if objnum > upload_count {
			break
		}
		prefix := fmt.Sprintf("%s/%s/Object-%d", url_host, bucket, objnum)
		req, _ := http.NewRequest("DELETE", prefix, nil)
		setSignature(req)
		if resp, err := httpClient.Do(req); err != nil {
			log.Fatalf("FATAL: Error deleting object %s: %v", prefix, err)
		} else if resp != nil && resp.StatusCode == http.StatusServiceUnavailable {
			atomic.AddInt32(&delete_slowdown_count, 1)
			atomic.AddInt32(&delete_count, -1)
		}
	}
	// Remember last done time
	delete_finish = time.Now()
	// One less thread
	atomic.AddInt32(&running_threads, -1)
}

func main() {
	// Hello
	fmt.Println("Wasabi benchmark program v2.0")

	// Parse command line
	myflag := flag.NewFlagSet("myflag", flag.ExitOnError)
	myflag.StringVar(&access_key, "a", "", "Access key")
	myflag.StringVar(&secret_key, "s", "", "Secret key")
	myflag.StringVar(&url_host, "u", "http://s3.wasabisys.com", "URL for host with method prefix")
	myflag.StringVar(&bucket, "b", "wasabi-benchmark-bucket", "Bucket for testing")
	myflag.StringVar(&region, "r", "US", "Region for testing")
	myflag.IntVar(&duration_secs, "d", 60, "Duration of each test in seconds")
	myflag.IntVar(&threads, "t", 1, "Number of threads to run")
	myflag.IntVar(&loops, "lp", 1, "Number of times to repeat test")
	myflag.IntVar(&localloopsleep, "ls", 0, "Number of times to repeat test")
	myflag.IntVar(&localloops, "l", 1, "Number of times to repeat test")
	myflag.IntVar(&filesize, "y", 1000, "bytes of object, unit bytes")
	myflag.BoolVar(&deleteobj, "c", false, "delete all obj")
	myflag.StringVar(&formKey, "k", "objkey", "objkey or prefix object name")
	fmt.Println("delete meesage", deleteobj)
	//formKey := "objkey"
	var sizeArg string
	myflag.StringVar(&sizeArg, "z", "1M", "Size of objects in bytes with postfix K, M, and G")
	if err := myflag.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	// Check the arguments
	if access_key == "" {
		log.Fatal("Missing argument -a for access key.")
	}
	if secret_key == "" {
		log.Fatal("Missing argument -s for secret key.")
	}
	var err error
	if object_size, err = bytefmt.ToBytes(sizeArg); err != nil {
		log.Fatalf("Invalid -z argument for object size: %v", err)
	}

	// Echo the parameters
	logit(fmt.Sprintf("Parameters: url=%s, bucket=%s, region=%s, duration=%d, threads=%d, loops=%d, size=%s",
		url_host, bucket, region, duration_secs, threads, localloops, filesize))

	if !deleteobj {
		// Initialize data for the bucket
		object_data = make([]byte, object_size)
		rand.Read(object_data)
		hasher := md5.New()
		hasher.Write(object_data)
		object_data_md5 = base64.StdEncoding.EncodeToString(hasher.Sum(nil))

		// Create the bucket and delete all the objects
		createBucket(true)
		//deleteAllObjects()

		// Loop running the tests

		//n := 10000
		n := filesize
		b := RandStringBytes(n)
		starttime := time.Now()
		wg.Add(int(threads))

		// reset counters
		upload_count = 0
		upload_slowdown_count = 0
		download_count = 0
		download_slowdown_count = 0
		delete_count = 0
		delete_slowdown_count = 0

		// Run the upload case
		running_threads = int32(threads)
		//endtime = starttime.Add(time.Second * time.Duration(duration_secs))
		for n := 0; n < threads; n++ {
			//go runUpload(n)
			//go runUploadNew(n, loop, bucket)
			go putObject(n, bucket, b, formKey)
		}

		// Wait for it to finish
		//fmt.Println("uploading now wait for finished")

		wg.Wait()
		/*
			upload_finish = time.Now()
			upload_time := upload_finish.Sub(starttime).Seconds()

			bps := float64(int(upload_count)*filesize*loop) / upload_time
			logit(fmt.Sprintf("Loop %d: PUT time %.1f secs, objects_count = %d, speed = %sB/sec, %.1f operations/sec. Slowdowns = %d",
				loop, upload_time, upload_count, bytefmt.ByteSize(uint64(bps)), float64(loop*threads)/upload_time, upload_slowdown_count))
		*/
		upload_finish = time.Now()
		end_time := upload_finish.Sub(starttime).Seconds()
		bps := float64(int(localloops*threads)*filesize) / end_time
		logresult(fmt.Sprintf("Finished: Loop %d: PUT time %.1f secs, objects_count = %d, concurrency = %d speed = %sB/sec, %.1f operations/sec. filesize = %d",
			localloops, end_time, localloops*threads, threads, bytefmt.ByteSize(uint64(bps)), float64(localloops*threads)/end_time, filesize))

	} else {
		fmt.Println("start to clena data")
		cleanAllObjects(localloops, threads, bucket, formKey)
	}
	/*
	 */
}
