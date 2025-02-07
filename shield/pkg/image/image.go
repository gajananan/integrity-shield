//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package image

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	ishieldconfig "github.com/open-cluster-management/integrity-shield/shield/pkg/config"
	cosigncli "github.com/sigstore/cosign/cmd/cosign/cli"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ImageVerifyResult struct {
	Object     unstructured.Unstructured `json:"object"`
	ImageRef   string                    `json:"imageRef"`
	Verified   bool                      `json:"verified"`
	InScope    bool                      `json:"inScope"`
	Signer     string                    `json:"signer"`
	SignedTime *time.Time                `json:"signedTime"`
	FailReason string                    `json:"failReason"`
}

type ImageVerifyOption struct {
	KeyPath string
}

// verify all images in a container of the specified resource
func VerifyImageInManifest(resource unstructured.Unstructured, profile ishieldconfig.ImageProfile) (bool, error) {
	yamlBytes, err := yaml.Marshal(resource.Object)
	if err != nil {
		return false, errors.Wrap(err, "failed to yaml.Marshal() the resource")
	}
	tmpDir, err := ioutil.TempDir("", "verify-image")
	if err != nil {
		return false, fmt.Errorf("failed to create temp dir: %s", err.Error())
	}
	defer os.RemoveAll(tmpDir)

	manifestPath := filepath.Join(tmpDir, "manifest.yaml")
	err = ioutil.WriteFile(manifestPath, yamlBytes, 0644)
	if err != nil {
		return false, fmt.Errorf("failed to create temp manifest file: %s", err.Error())
	}

	keyPathList := []string{}
	for _, keyConfig := range profile.KeyConfigs {
		if keyConfig.KeySecretName != "" {
			keyPath, err := ishieldconfig.LoadKeySecret(keyConfig.KeySecretNamespace, keyConfig.KeySecretName)
			if err != nil {
				return false, errors.Wrap(err, "failed to load a key secret for image verification")
			}
			keyPathList = append(keyPathList, keyPath)
		}
	}
	if len(keyPathList) == 0 {
		keyPathList = []string{""} // for keyless verification
	}

	allImagesVerified := false
	failReason := ""
	// overallFailReason := ""
	for _, keyPath := range keyPathList {
		cmd := cosigncli.VerifyManifestCommand{VerifyCommand: cosigncli.VerifyCommand{}}
		if keyPath != "" {
			cmd.KeyRef = keyPath
		}

		var verifiedWithThisKey bool
		// currently cosigncli.VerifyManifestCommand.Exec() does not return detail information like image names and their signer names
		// TODO: create an issue in sigstore/cosign for this function to return some additional information
		iErr := cmd.Exec(context.Background(), []string{manifestPath})
		if iErr == nil {
			verifiedWithThisKey = true
		} else {
			failReason = iErr.Error()
		}
		if verifiedWithThisKey {
			allImagesVerified = true
			break
		}
	}
	var retErr error
	if !allImagesVerified {
		retErr = errors.New(failReason)
	}

	return allImagesVerified, retErr
}
