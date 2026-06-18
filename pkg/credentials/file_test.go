/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2017 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package credentials

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFileAWS(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("\"/bin/cat\": file does not exist")
	}
	os.Clearenv()

	creds := NewFileAWSCredentials("credentials.sample", "")
	credValues, err := creds.GetWithContext(defaultCredContext)
	if err != nil {
		t.Fatal(err)
	}

	if credValues.AccessKeyID != "accessKey" {
		t.Errorf("Expected 'accessKey', got %s'", credValues.AccessKeyID)
	}
	if credValues.SecretAccessKey != "secret" {
		t.Errorf("Expected 'secret', got %s'", credValues.SecretAccessKey)
	}
	if credValues.SessionToken != "token" {
		t.Errorf("Expected 'token', got %s'", credValues.SessionToken)
	}

	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", "credentials.sample")
	creds = NewFileAWSCredentials("", "")
	credValues, err = creds.GetWithContext(defaultCredContext)
	if err != nil {
		t.Fatal(err)
	}

	if credValues.AccessKeyID != "accessKey" {
		t.Errorf("Expected 'accessKey', got %s'", credValues.AccessKeyID)
	}
	if credValues.SecretAccessKey != "secret" {
		t.Errorf("Expected 'secret', got %s'", credValues.SecretAccessKey)
	}
	if credValues.SessionToken != "token" {
		t.Errorf("Expected 'token', got %s'", credValues.SessionToken)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(wd, "credentials.sample"))
	creds = NewFileAWSCredentials("", "")
	credValues, err = creds.GetWithContext(defaultCredContext)
	if err != nil {
		t.Fatal(err)
	}

	if credValues.AccessKeyID != "accessKey" {
		t.Errorf("Expected 'accessKey', got %s'", credValues.AccessKeyID)
	}
	if credValues.SecretAccessKey != "secret" {
		t.Errorf("Expected 'secret', got %s'", credValues.SecretAccessKey)
	}
	if credValues.SessionToken != "token" {
		t.Errorf("Expected 'token', got %s'", credValues.SessionToken)
	}

	os.Clearenv()
	t.Setenv("AWS_PROFILE", "no_token")

	creds = NewFileAWSCredentials("credentials.sample", "")
	credValues, err = creds.GetWithContext(defaultCredContext)
	if err != nil {
		t.Fatal(err)
	}

	if credValues.AccessKeyID != "accessKey" {
		t.Errorf("Expected 'accessKey', got %s'", credValues.AccessKeyID)
	}
	if credValues.SecretAccessKey != "secret" {
		t.Errorf("Expected 'secret', got %s'", credValues.SecretAccessKey)
	}

	os.Clearenv()

	creds = NewFileAWSCredentials("credentials.sample", "no_token")
	credValues, err = creds.GetWithContext(defaultCredContext)
	if err != nil {
		t.Fatal(err)
	}

	if credValues.AccessKeyID != "accessKey" {
		t.Errorf("Expected 'accessKey', got %s'", credValues.AccessKeyID)
	}
	if credValues.SecretAccessKey != "secret" {
		t.Errorf("Expected 'secret', got %s'", credValues.SecretAccessKey)
	}

	creds = NewFileAWSCredentials("credentials-non-existent.sample", "no_token")
	_, err = creds.GetWithContext(defaultCredContext)
	if !os.IsNotExist(err) {
		t.Errorf("Expected open non-existent.json: no such file or directory, got %s", err)
	}
	if !creds.IsExpired() {
		t.Error("Should be expired if not loaded")
	}

	os.Clearenv()

	creds = NewFileAWSCredentials("credentials.sample", "with_process")
	credValues, err = creds.GetWithContext(defaultCredContext)
	if err != nil {
		t.Fatal(err)
	}

	if credValues.AccessKeyID != "accessKey" {
		t.Errorf("Expected 'accessKey', got %s'", credValues.AccessKeyID)
	}
	if credValues.SecretAccessKey != "secret" {
		t.Errorf("Expected 'secret', got %s'", credValues.SecretAccessKey)
	}
	if credValues.SessionToken != "token" {
		t.Errorf("Expected 'token', got %s'", credValues.SessionToken)
	}
	if creds.IsExpired() {
		t.Error("Should not be expired")
	}
}
