/*
Copyright © 2022 Miroslav Safar <msafar@redhat.com>

*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	registryinstance "github.com/redhat-developer/app-services-sdk-go/registryinstance/apiv1internal"
	registryinstanceclient "github.com/redhat-developer/app-services-sdk-go/registryinstance/apiv1internal/client"

	"github.com/spf13/cobra"
)

type SyncOptions struct {
	registryUrl string

	srcDir string

	serviceRegistryApi *registryinstanceclient.APIClient
}

type ArtifactInfo struct {
	ArtifactType registryinstanceclient.ArtifactType `json:"artifactType"`
}

func NewSyncCommand() *cobra.Command {
	opts := &SyncOptions{}

	cmd := &cobra.Command{
		Use:     "sync",
		Short:   "",
		Long:    "",
		Example: "",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			return runSync(opts)
		},
	}

	cmd.Flags().StringVar(&opts.srcDir, "src", "", "Source directory")
	cmd.Flags().StringVar(&opts.registryUrl, "registryUrl", "", "Registry url")

	return cmd
}

func runSync(opts *SyncOptions) error {
	baseURL := opts.registryUrl + "/apis/registry/v2"
	cfg := &registryinstance.Config{
		Debug:   false,
		BaseURL: baseURL,
	}

	client := registryinstance.NewAPIClient(cfg)

	opts.serviceRegistryApi = client

	if opts.srcDir == "" {
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}
		opts.srcDir = path
	}

	files, err := os.ReadDir(opts.srcDir)
	if err != nil {
		return err
	}

	for _, element := range files {
		if element.IsDir() {
			err = syncGroup(opts, element.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func syncGroup(opts *SyncOptions, groupId string) error {
	fmt.Println("Syncing artifacts of group " + groupId)
	artifacts, err := os.ReadDir(opts.srcDir + "/" + groupId)
	if err != nil {
		return err
	}

	for _, artifactEl := range artifacts {
		if artifactEl.IsDir() {
			err = syncArtifact(opts, groupId, artifactEl.Name())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func syncArtifact(opts *SyncOptions, groupId string, artifactId string) error {
	fmt.Println("Syncing artifact " + groupId + ":" + artifactId)

	artifactInfoStr, err := os.ReadFile(opts.srcDir + "/" + groupId + "/" + artifactId + "/artifact.json")
	if err != nil {
		return err
	}

	var artifactInfo ArtifactInfo
	err = json.Unmarshal(artifactInfoStr, &artifactInfo)
	if err != nil {
		return err
	}

	versions, err := os.ReadDir(opts.srcDir + "/" + groupId + "/" + artifactId)
	if err != nil {
		return err
	}

	for _, versionEl := range versions {
		if versionEl.IsDir() {
			err = syncArtifactVersion(opts, groupId, artifactId, artifactInfo.ArtifactType, versionEl.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func syncArtifactVersion(opts *SyncOptions, groupId string, artifactId string, artifactType registryinstanceclient.ArtifactType, versionId string) error {
	fmt.Println("Syncing artifact version " + groupId + ":" + artifactId + ":" + versionId)

	content, err := os.ReadFile(opts.srcDir + "/" + groupId + "/" + artifactId + "/" + versionId + "/content.data")
	if err != nil {
		return err
	}

	var references []registryinstanceclient.ArtifactReference
	referencesJson, err := os.ReadFile(opts.srcDir + "/" + groupId + "/" + artifactId + "/" + versionId + "/references.json")
	if err == nil {
		err = json.Unmarshal(referencesJson, &references)
		if err != nil {
			return err
		}
	} else {
		references = make([]registryinstanceclient.ArtifactReference, 0)
	}

	req := opts.serviceRegistryApi.ArtifactsApi.CreateArtifact(context.Background(), groupId)
	req = req.IfExists(registryinstanceclient.IFEXISTS_RETURN_OR_UPDATE)
	req = req.XRegistryArtifactId(artifactId)
	req = req.XRegistryArtifactType(artifactType)
	req = req.XRegistryVersion(versionId)
	req = req.Body(string(content))
	// TODO: Report bug in SDK - cannot create content with references because of wrong content type
	//req = req.Body(&registryinstanceclient.ContentCreateRequest{
	//	Content:    string(content),
	//	References: references,
	//})

	_, _, err = req.Execute()
	if err != nil {
		fmt.Print(err)
		return err
	}

	return nil
}
