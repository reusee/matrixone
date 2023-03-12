package fileservice

import (
	"github.com/matrixorigin/matrixone/pkg/util/metric"
	"go.uber.org/zap"
)

type CachingFsStatsCollector struct {
	cachingFs *CachingFileService
}

func NewCachingFsStatsCollector(cachingFs *CachingFileService) metric.StatsCollector {
	return &CachingFsStatsCollector{
		cachingFs: cachingFs,
	}
}

// Collect returns the fields and its values.
// TODO: Replace []zap.Field with `map[string]int` or `Struct`
func (c *CachingFsStatsCollector) Collect() []zap.Field {
	var fields []zap.Field

	counter := (*c.cachingFs).CacheCounter()
	fields = append(fields, zap.Any("S3ListObjects", counter.S3ListObjects))
	fields = append(fields, zap.Any("S3HeadObject", counter.S3HeadObject))
	fields = append(fields, zap.Any("S3PutObject", counter.S3PutObject))
	fields = append(fields, zap.Any("S3GetObject", counter.S3GetObject))
	fields = append(fields, zap.Any("S3DeleteObjects", counter.S3DeleteObjects))
	fields = append(fields, zap.Any("S3DeleteObject", counter.S3DeleteObject))

	fields = append(fields, zap.Any("CacheRead", counter.CacheRead))
	fields = append(fields, zap.Any("CacheHit", counter.CacheHit))

	fields = append(fields, zap.Any("MemCacheRead", counter.MemCacheRead))
	fields = append(fields, zap.Any("MemCacheHit", counter.MemCacheHit))

	fields = append(fields, zap.Any("DiskCacheRead", counter.DiskCacheRead))
	fields = append(fields, zap.Any("DiskCacheHit", counter.DiskCacheHit))

	return fields
}
