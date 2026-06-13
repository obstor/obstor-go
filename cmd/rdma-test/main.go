// Copyright 2024-2026 - MinIO, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// rdma-test exercises the unified obstor-go RDMA path against a running
// Obstor server. Requires -tags=rdma at build time and libobstorcpp.so on
// the host's library search path.
//
//   go build -tags=rdma -o rdma-test ./cmd/rdma-test
//   OBSTOR_ENDPOINT=coe01:9000 OBSTOR_ACCESS_KEY=... OBSTOR_SECRET_KEY=... ./rdma-test

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"unsafe"

	obstor "github.com/obstor/obstor-go/v7"
	"github.com/obstor/obstor-go/v7/pkg/credentials"
)

const (
	testBucket = "rdma-test"
	testObject = "test-object-cpu"
	testSize   = 1 << 20 // 1 MiB
)

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	endpoint := envOr("OBSTOR_ENDPOINT", "coe01:9000")
	accessKey := envOr("OBSTOR_ACCESS_KEY", "obstoradmin")
	secretKey := envOr("OBSTOR_SECRET_KEY", "obstoradmin")

	fmt.Printf("endpoint=%s rdma_available=%v\n", endpoint, obstor.IsRDMAAvailable())

	client, err := obstor.New(endpoint, &obstor.Options{
		Creds:      credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:     false,
		EnableRDMA: true,
	})
	if err != nil {
		return fmt.Errorf("New: %w", err)
	}

	ctx := context.Background()

	exists, err := client.BucketExists(ctx, testBucket)
	if err != nil {
		return fmt.Errorf("BucketExists: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, testBucket, obstor.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("MakeBucket: %w", err)
		}
	}

	src := obstor.AlignedBuffer(testSize)
	if src == nil {
		return fmt.Errorf("AlignedBuffer(%d) returned nil", testSize)
	}
	defer obstor.FreeAlignedBuffer(src)
	srcSlice := unsafe.Slice((*byte)(src), testSize)
	for i := range srcSlice {
		srcSlice[i] = byte(i)
	}

	fmt.Print("PutObject (RDMA)... ")
	info, err := client.PutObject(ctx, testBucket, testObject, nil, 0, obstor.PutObjectOptions{
		RDMABuffer:     src,
		RDMABufferSize: testSize,
	})
	if err != nil {
		return fmt.Errorf("PutObject: %w", err)
	}
	fmt.Printf("ok etag=%s size=%d checksum=%s\n", info.ETag, info.Size, info.ChecksumCRC64NVME)

	dst := obstor.AlignedBuffer(testSize)
	if dst == nil {
		return fmt.Errorf("AlignedBuffer(%d) returned nil", testSize)
	}
	defer obstor.FreeAlignedBuffer(dst)
	dstSlice := unsafe.Slice((*byte)(dst), testSize)

	fmt.Print("GetObject (RDMA)... ")
	obj, err := client.GetObject(ctx, testBucket, testObject, obstor.GetObjectOptions{
		RDMABuffer:     dst,
		RDMABufferSize: testSize,
	})
	if err != nil {
		return fmt.Errorf("GetObject: %w", err)
	}
	stat, err := obj.Stat()
	if err != nil {
		return fmt.Errorf("Stat: %w", err)
	}
	fmt.Printf("ok size=%d\n", stat.Size)

	if !bytes.Equal(srcSlice, dstSlice) {
		return fmt.Errorf("FAIL: roundtrip data mismatch")
	}
	fmt.Println("PASS: roundtrip verified")
	return nil
}
