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

package helm

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/IBM/integrity-enforcer/enforcer/pkg/cache"
	"github.com/IBM/integrity-enforcer/enforcer/pkg/kubeutil"
	logger "github.com/IBM/integrity-enforcer/enforcer/pkg/logger"
	"github.com/IBM/integrity-enforcer/enforcer/pkg/mapnode"
	sign "github.com/IBM/integrity-enforcer/enforcer/pkg/sign"
	"helm.sh/helm/v3/pkg/provenance"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	v1cli "k8s.io/client-go/kubernetes/typed/core/v1"
)

const releaseSecretPrefix = "sh.helm.release.v1"
const releaseSecretType = "helm.sh/release.v1"

type ReleaseObject struct {
	Data *release.Release `json:"data"`
}

func NewReleaseObject(data *release.Release) *ReleaseObject {
	return &ReleaseObject{
		Data: data,
	}
}

type PackageInfo struct {
	FileUrl  string
	FilePath string
	ProvUrl  string
	ProvPath string
}

type ValuesInfo struct {
	Message     string
	Signature   string
	EmptyConfig bool
}

type HelmInfo struct {
	ChartName    string
	ChartVersion string
	Package      PackageInfo
	Values       ValuesInfo
}

func DecodeReleaseSecretFromRawBytes(rawBytes []byte) *ReleaseObject {

	var releaseSecret *v1.Secret
	err := json.Unmarshal(rawBytes, &releaseSecret)
	if err != nil {
		logger.Warn("Error when unmarshaling request object", err)
		return nil
	}

	releaseString := string(releaseSecret.Data["release"])
	if releaseString == "" {
		return nil
	}
	if rls, err := decodeRelease(releaseString); err != nil {
		logger.Warn("Error when decoding release object", err)
		return nil
	} else {
		return NewReleaseObject(rls)
	}
}

func GetPackageInfo(rawBytes []byte, chartRepo, chartDir string) (*HelmInfo, error) {
	rlsObj := DecodeReleaseSecretFromRawBytes(rawBytes)
	rls := rlsObj.Data
	emptyConfig := (rls.Config == nil)
	chartName := rls.Chart.Metadata.Name
	chartVersion := rls.Chart.Metadata.Version
	pkgFileName := fmt.Sprintf("%s-%s.tgz", chartName, chartVersion)
	pkgProvName := fmt.Sprintf("%s-%s.tgz.prov", chartName, chartVersion)
	pkgFilePath := path.Join(chartDir, pkgFileName)
	pkgProvPath := path.Join(chartDir, pkgProvName)
	pkgFileUrl := fmt.Sprintf("%s/%s", chartRepo, pkgFileName)
	pkgProvUrl := fmt.Sprintf("%s/%s", chartRepo, pkgProvName)
	_, err := getChartFiles(pkgFileUrl, pkgProvUrl, pkgFilePath, pkgProvPath)
	if err != nil {
		return nil, err
	}
	packageInfo := PackageInfo{
		FileUrl:  pkgFileUrl,
		FilePath: pkgFilePath,
		ProvUrl:  pkgProvUrl,
		ProvPath: pkgProvPath,
	}
	valuesInfo := ValuesInfo{
		Message:     "",
		Signature:   "",
		EmptyConfig: emptyConfig,
	}
	return &HelmInfo{
		ChartName:    chartName,
		ChartVersion: chartVersion,
		Package:      packageInfo,
		Values:       valuesInfo,
	}, nil
}

func VerifyPackage(filePath, provPath, keyringPath string) (*sign.Signer, error) {
	sig, err := provenance.NewFromKeyring(keyringPath, "")
	if err != nil {
		return nil, err
	}
	veri, _ := sig.Verify(filePath, provPath)
	if veri.SignedBy == nil {
		return nil, nil
	}
	signIdt := sign.GetFirstIdentity(veri.SignedBy)
	signer := &sign.Signer{
		Email:   signIdt.UserId.Email,
		Name:    signIdt.UserId.Name,
		Comment: signIdt.UserId.Comment,
	}
	return signer, nil
}

func decodeRelease(data string) (*release.Release, error) {
	magicGzip := []byte{0x1f, 0x8b, 0x08}
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(b[0:3], magicGzip) {
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		b2, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		b = b2
	}

	var rls release.Release
	if err := json.Unmarshal(b, &rls); err != nil {
		return nil, err
	}
	return &rls, nil
}

func getChartFiles(pkgFileUrl, pkgProvUrl, pkgFilePath, pkgProvPath string) (bool, error) {
	fileCached := (cache.GetString(pkgFileUrl) == pkgFilePath)
	if fileCached {
		// pass
	} else {
		err := DownloadFile(pkgFileUrl, pkgFilePath)
		if err != nil {
			logger.Error(err)
			return false, err
		}
		cache.Set(pkgFileUrl, pkgFilePath)
	}

	provCached := (cache.GetString(pkgProvUrl) == pkgProvPath)
	if provCached {
		// pass
	} else {
		err := DownloadFile(pkgProvUrl, pkgProvPath)
		if err != nil {
			logger.Error(err)
			return false, err
		}
		cache.Set(pkgProvUrl, pkgProvPath)
	}
	return true, nil
}

func FindReleaseSecret(namespace, kind, name string, rawObj []byte, isReleaseSecret bool) ([]byte, error) {
	config, err := kubeutil.GetKubeConfig()
	if err != nil {
		return nil, err
	}
	var rsec *v1.Secret
	if isReleaseSecret {
		err = json.Unmarshal(rawObj, &rsec)
		if err != nil {
			return nil, err
		}
	} else {
		v1client := v1cli.NewForConfigOrDie(config)
		rsecList, err := v1client.Secrets(namespace).List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, rseci := range rsecList.Items {
			if rseci.Type != releaseSecretType {
				continue
			}
			if ok, mnfObj := findObjNameInReleaseSecret(kind, name, &rseci); ok {
				if matchWithManifest(rawObj, mnfObj) {
					rsec = &rseci
				}
			}
		}
	}
	if rsec == nil {
		return nil, nil
	}
	rsecBytes, err := json.Marshal(rsec)
	if err != nil {
		return nil, err
	}
	return rsecBytes, nil
}

func findObjNameInReleaseSecret(kind, name string, rsec *v1.Secret) (bool, []byte) {
	rsecBytes, err := json.Marshal(rsec)
	if err != nil {
		logger.Error(err)
		return false, nil
	}
	rls := DecodeReleaseSecretFromRawBytes(rsecBytes)
	if rls == nil {
		logger.Error("release secret is nil")
		return false, nil
	}

	manifestMap := releaseutil.SplitManifests(rls.Data.Manifest)

	for _, mYaml := range manifestMap {
		mJson, err := yaml.ToJSON([]byte(mYaml))
		if err != nil {
			logger.Error(err)
			continue
		}
		u := &unstructured.Unstructured{}
		err = u.UnmarshalJSON([]byte(mJson))
		if err != nil {
			logger.Error(err)
			continue
		}

		if u.GetName() == name && u.GetKind() == kind {
			return true, mJson
		}
	}
	return false, nil
}

func matchWithManifest(requestObject, manifestObject []byte) bool {
	reqNode, err := mapnode.NewFromBytes(requestObject)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to load request object into mapnode; %s", err.Error()))
	}
	mnfNode, err := mapnode.NewFromBytes(manifestObject)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to load request object into mapnode; %s", err.Error()))
	}
	diff := mnfNode.FindUpdatedAndDeleted(reqNode)
	matched := (diff.Size() == 0)
	if !matched {
		logger.Debug(fmt.Sprintf("Diff found in matching with manifest; %s", diff.ToJson()))
	}
	return matched
}

func IsReleaseSecret(kind, name string) bool {
	return (kind == "Secret" && strings.HasPrefix(name, releaseSecretPrefix))
}
