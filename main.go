package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/logutils"
)

var version string

const (
	ENV_SRC       = "CONTAINER_EXEC_SRC"
	ENV_CODE_DIR  = "CONTAINER_EXEC_CODE_DIR"
	ENV_EVENT     = "CONTAINER_EXEC_EVENT"
	ENV_LOG_LEVEL = "CONTAINER_EXEC_LOG_LEVEL"

	DEFAULT_CODE_DIR = "/tmp/lambda"
)

type Event interface{}
type Result interface{}

func main() {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(os.Getenv(ENV_LOG_LEVEL)),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	var (
		versionFlag bool
	)

	flag.BoolVar(&versionFlag, "version", false, "show version")
	flag.Parse()

	if versionFlag {
		log.Printf("lambda-container-exec v%s", version)
		os.Exit(0)
	}

	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, event Event) (Result, error) {
	rawEvent, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("failed to unmarshal event=%#v error=%s", event, err)
	}

	// place code

	bucket, key, err := parseS3Path(os.Getenv(ENV_SRC))
	if err != nil {
		log.Printf("[DEBUG] failed to parseS3Path")
		return nil, err
	}

	codeDir := os.Getenv(ENV_CODE_DIR)
	if codeDir == "" {
		codeDir = DEFAULT_CODE_DIR
	}

	funcDir := filepath.Join(codeDir, key)
	if err := os.MkdirAll(funcDir, 0755); err != nil {
		log.Printf("[DEBUG] failed to create func dir '%s'", funcDir)
		return nil, err
	}

	if err := placeSourceCode(ctx, bucket, key, funcDir); err != nil {
		log.Printf("[DEBUG] failed to place source code")
		return nil, err
	}

	// exec code

	envVars := os.Environ()
	envVars = append(envVars, fmt.Sprintf("%s=%s", ENV_EVENT, rawEvent))

	bootstrapPath := filepath.Join(funcDir, "bootstrap")
	bootstrap, err := exec.LookPath(bootstrapPath)
	if err != nil {
		log.Printf("[WARN] bootstrap not found at %s", bootstrapPath)
		return nil, err
	}

	log.Printf("[INFO] running bootstrap ...")
	return runCmd(ctx, bootstrap, envVars)
}

func runCmd(ctx context.Context, cmdPath string, envVars []string) ([]byte, error) {
	log.Printf("[INFO] command path=%s", cmdPath)

	cmd := exec.Command(cmdPath)
	cmd.Env = envVars

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[WARN] failed to get stdout pipe error='%s'", err)
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("[WARN] failed to get stderr pipe error='%s'", err)
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[WARN] failed to start command error='%s'", err)
		return nil, err
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	out, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Printf("[WARN] failed to read command output error='%s", err)
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("[WARN] failed to wait command error='%s'", err)
		return nil, err
	}

	log.Printf("[INFO] done command")

	return out, nil
}

func placeSourceCode(ctx context.Context, bucket, key string, dest string) error {
	log.Printf("[DEBUG] start to place source code from bucket=%s key=%s", bucket, key)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("[DEBUG] failed to load aws config")
		return err
	}

	client := s3.NewFromConfig(cfg)

	in := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	out, err := client.GetObject(ctx, in)
	if err != nil {
		log.Printf("[DEBUG] failed to GetObject: %s", err)
		return err
	}

	if err := unarchiveTarball(out.Body, dest); err != nil {
		log.Printf("[DEBUG] failed to unarchive tarball")
		return err
	}

	log.Printf("[DEBUG] tarball unarchived to %s", dest)
	log.Printf("[DEBUG] end to place source code")

	return nil
}

// parseS3Path returns bucket name and object key.
func parseS3Path(path string) (string, string, error) {
	u, err := url.Parse(path)
	if err != nil {
		log.Printf("[DEBUG] failed to parse S3 path as URL")
		return "", "", err
	}

	if u.Scheme != "s3" {
		return "", "", fmt.Errorf("not a S3 path")
	}

	return u.Host, strings.TrimPrefix(u.Path, "/"), nil
}

func unarchiveTarball(r io.Reader, dest string) error {
	log.Printf("[DEBUG] start to unarchiving tarball")

	gr, err := gzip.NewReader(r)
	if err != nil {
		log.Printf("[DEBUG] failed to create gzip reader")
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			log.Printf("[DEBUG] end to unarchiving tarball")
			return nil
		case err != nil:
			log.Printf("[DEBUG] something wrong on unarchive tarball")
			return err
		case header == nil:
			continue
		}

		log.Printf("[DEBUG] processing %s", header.Name)

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.Mkdir(target, 0755); err != nil {
					log.Printf("[DEBUG] failed to mkdir %s", target)
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				log.Printf("[DEBUG] failed to os.OpenFile %s", target)
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				log.Printf("[DEBUG] failed to os.Copy %s", target)
				return err
			}

			f.Close()
		}
	}
}
