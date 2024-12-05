// Copyright 2024 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fileservice

import (
	"context"
	"io"
	"iter"
	"time"

	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/perfcounter"
	"go.uber.org/zap"
)

type HDFSObjectStorage struct {
	perfCounterSets []*perfcounter.CounterSet
}

func NewHDFSObjectStorage(
	_ context.Context,
	args ObjectStorageArguments,
	perfCounterSets []*perfcounter.CounterSet,
) (*HDFSObjectStorage, error) {

	if err := args.validate(); err != nil {
		return nil, err
	}

	logutil.Info("new object storage",
		zap.Any("sdk", "hdfs"),
		zap.Any("arguments", args),
	)

	//TODO client

	return &HDFSObjectStorage{
		perfCounterSets: perfCounterSets,
	}, nil
}

var _ ObjectStorage = new(HDFSObjectStorage)

func (d *HDFSObjectStorage) Delete(ctx context.Context, keys ...string) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}

	perfcounter.Update(ctx, func(counter *perfcounter.CounterSet) {
		counter.FileService.S3.Delete.Add(1)
	}, d.perfCounterSets...)

	for _, key := range keys {
		//TODO remove
		_ = key
	}

	return nil
}

func (d *HDFSObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	//TODO
	return false, nil
}

func (d *HDFSObjectStorage) List(
	ctx context.Context,
	prefix string,
) iter.Seq2[*DirEntry, error] {
	return func(yield func(*DirEntry, error) bool) {
		if err := ctx.Err(); err != nil {
			yield(nil, err)
			return
		}

		perfcounter.Update(ctx, func(counter *perfcounter.CounterSet) {
			counter.FileService.S3.List.Add(1)
		}, d.perfCounterSets...)

		//TODO

	}
}

func (d *HDFSObjectStorage) Read(ctx context.Context, key string, min *int64, max *int64) (r io.ReadCloser, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	perfcounter.Update(ctx, func(counter *perfcounter.CounterSet) {
		counter.FileService.S3.Get.Add(1)
	}, d.perfCounterSets...)

	//TODO

	return nil, nil
}

func (d *HDFSObjectStorage) Stat(ctx context.Context, key string) (size int64, err error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	perfcounter.Update(ctx, func(counter *perfcounter.CounterSet) {
		counter.FileService.S3.Head.Add(1)
	}, d.perfCounterSets...)

	//TODO

	return
}

func (d *HDFSObjectStorage) Write(ctx context.Context, key string, r io.Reader, sizeHint *int64, expire *time.Time) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}

	perfcounter.Update(ctx, func(counter *perfcounter.CounterSet) {
		counter.FileService.S3.Put.Add(1)
	}, d.perfCounterSets...)

	//TODO

	return nil
}
