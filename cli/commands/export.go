package commands

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	v2 "github.com/urfave/cli/v2"

	"gopkg.in/yaml.v2"

	"github.com/cmattoon/dockerenv/pkg/inspector"
	//"golang.org/x/crypto/ssh/terminal"
)

var log *logrus.Logger
var S3_BUCKET string
var OUTPUT_DIR string
var S3_REGION string

func init() {
	log = logrus.New()
	S3_BUCKET, _ = os.LookupEnv("S3_BUCKET")
	OUTPUT_DIR, _ = os.LookupEnv("OUTPUT_DIR")
	if r, ok := os.LookupEnv("S3_REGION"); ok && r != "" {
		S3_REGION = r
	} else if r, ok := os.LookupEnv("AWS_DEFAULT_REGION"); ok && r != "" {
		S3_REGION = r
	}
}

func ExportCommand() *v2.Command {
	return &v2.Command{
		Name:   "export",
		Usage:  "exports the current environment",
		Action: exportContainerEnvAction,
		Flags: []v2.Flag{
			&v2.StringFlag{
				Name:  "format",
				Usage: "The output format (yaml, env, json, ssm, s3)",
			},
			&v2.StringFlag{
				Name:  "path-prefix",
				Usage: "The SSM or S3 path prefix. Should start with /",
			},
			&v2.StringFlag{
				Name:  "container-id",
				Usage: "The container ID (default: ALL)",
			},
			&v2.StringFlag{
				Name:  "s3-bucket",
				Usage: "The S3 Bucket name (no protocol)",
			},
			&v2.StringFlag{
				Name:  "output-dir",
				Usage: "output directory",
			},
			&v2.BoolFlag{
				Name:  "snapshot",
				Usage: "Also writes a set of params useful for restarting the container",
			},
			&v2.BoolFlag{
				Name:  "overwrite",
				Usage: "Set this to overwrite an existing set of files",
			},
		},
	}
}

type ContainerInfo struct {
	Name   string
	Image  string
	Cmd    []string
	Env    []string
	Labels map[string]string
}

func newContainerInfo(container types.ContainerJSON) ContainerInfo {
	containerInfo := ContainerInfo{
		Name:   container.Name,
		Image:  container.Config.Image,
		Cmd:    container.Config.Cmd,
		Env:    container.Config.Env,
		Labels: make(map[string]string),
	}
	for key, value := range container.Config.Labels {
		containerInfo.Labels[key] = value
	}
	return containerInfo
}

func exportContainerEnvAction(c *v2.Context) error {
	pathPrefix := c.String("path-prefix")
	if len(pathPrefix) < 1 || !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/"
	}
	outputDir := c.String("output-dir")
	if outputDir != "" {
		OUTPUT_DIR = outputDir
	}
	log.Infof("Using output directory at %s", OUTPUT_DIR)
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	var containerId string
	if cid := c.String("container-id"); cid != "" {
		containerId = cid
	} else {
		containerId = "ALL"
	}

	// Get list of containers
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	allValues := map[string]map[string]string{}

	ins, err := inspector.New()
	if err != nil {
		return fmt.Errorf("failed to create docker inspector: %w", err)
	}

	snapshot := c.Bool("snapshot")
	metaFiles := map[string][]byte{}
	for _, container := range containers {
		if containerId == "ALL" || containerId == container.ID {
			if snapshot {
				c, err := cli.ContainerInspect(context.Background(), container.ID)
				info := newContainerInfo(c)
				log.Debug(info)
				meta, err := yaml.Marshal(info)
				if err != nil {
					return fmt.Errorf("Failed to marshal YAML: %w", err)
				}

				metaFile := pathPrefix + "/containers/" + container.ID + "/container-meta.yaml"
				// if err = writeFileData(metaFile, meta); err != nil {
				// 	return fmt.Errorf("failed to write meta file data to %s: %w", metaFile, err)
				// }
				metaFiles[metaFile] = meta
			}

			values, err := ins.GetAllValues(container.ID)
			if err != nil {
				log.Error(err)
				continue
			}
			shortID := container.ID[0:8]
			allValues[shortID] = values
		}
	}

	format := c.String("format")
	if format == "" {
		format = "env"
	}

	switch format {
	case "json":
		return fmt.Errorf("finish me")

	case "yaml":
		data, err := yaml.Marshal(allValues)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		return ioutil.WriteFile("output.yaml", data, 0644)
	case "ssm":
		for cid, cenv := range allValues {
			for k, v := range cenv {
				ssmPath := fmt.Sprintf("%s/%s/%s", pathPrefix, cid, k)
				log.Infof("Saving \033[33m%s\033[0m as \033[36m%s\033[0m", ssmPath, v)
			}
		}
	case "env", "s3":

		var txt strings.Builder
		for cid, cenv := range allValues {

			for key, val := range cenv {
				txt.WriteString(fmt.Sprintf("%s=\"%s\"\n", key, val))
			}
			containersPrefix := pathPrefix + "/containers"
			containerPrefix := containersPrefix + "/" + cid
			envFileName := containerPrefix + "/container.env"

			if format == "env" {
				if err := writeFileData(envFileName, []byte(txt.String())); err != nil {
					log.Errorf("failed to write to %s: %s", OUTPUT_DIR+envFileName, err)
				}
			} else if format == "s3" {
				containersPrefix = pathPrefix + "/containers"
				containerPrefix = containersPrefix + "/" + cid[0:8]
				envFileName = containerPrefix + "/container.env"

				if err := writeS3Data(envFileName, []byte(txt.String())); err != nil {
					log.Errorf("failed to write to S3: %s", err)
				}
			}
		}
		for name, data := range metaFiles {
			if format == "s3" {
				if err := writeS3Data(name, data); err != nil {
					log.Errorf("failed to write to S3: %s", err)
				}
			}
		}
	}
	return fmt.Errorf("not implemented")
}

func writeFileData(filename string, data []byte) error {
	final_filename := OUTPUT_DIR + filename // prefix + filename
	log.Infof("Writing %d bytes to %s", len(data), final_filename)
	return ioutil.WriteFile(final_filename, data, 0644)
}

func writeS3Data(filename string, data []byte) error {
	var dryRun bool
	if S3_BUCKET == "" {
		dryRun = true
		S3_BUCKET = "S3_BUCKET"
		log.Warningf("Treating this as a dry run because S3_BUCKET Is not set")
	}

	s3FullPath := fmt.Sprintf("s3://%s%s", S3_BUCKET, filename)
	log.Infof("Writing %d bytes to %s", len(data), s3FullPath)
	log.Debugf("\033[33m%s\033[0m", data)
	if dryRun {
		log.Warningf("S3_BUCKET is empty. No data will be written")
		return nil // use as dry run for now
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(S3_REGION),
	})
	if err != nil {
		return err
	}

	s3client := s3.New(sess)

	// Write data to S3
	input := &s3.PutObjectInput{
		Bucket: aws.String(S3_BUCKET),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(data),
	}
	_, err = s3client.PutObject(input)
	if err != nil {
		return err
	}

	log.Infof("Wrote %d bytes to %s\n", len(data), s3FullPath)

	return nil
}
