package stats

import (
	"go.uber.org/zap"
)

const CachingFsFamilyName = "CachingFileService"

var familyNames = []string{
	CachingFsFamilyName,
}

var S3ListObjects = NewStatsCounter(CachingFsFamilyName, "S3ListObjects")
var S3HeadObject = NewStatsCounter(CachingFsFamilyName, "S3HeadObject")
var S3PutObject = NewStatsCounter(CachingFsFamilyName, "S3PutObject")
var S3GetObject = NewStatsCounter(CachingFsFamilyName, "S3GetObject")
var S3DeleteObjects = NewStatsCounter(CachingFsFamilyName, "S3DeleteObjects")
var S3DeleteObject = NewStatsCounter(CachingFsFamilyName, "S3DeleteObject")

var CacheRead = NewStatsCounter(CachingFsFamilyName, "CacheRead")
var CacheHit = NewStatsCounter(CachingFsFamilyName, "CacheHit")
var MemCacheRead = NewStatsCounter(CachingFsFamilyName, "MemCacheRead")
var MemCacheHit = NewStatsCounter(CachingFsFamilyName, "MemCacheHit")
var DiskCacheRead = NewStatsCounter(CachingFsFamilyName, "DiskCacheRead")
var DiskCacheHit = NewStatsCounter(CachingFsFamilyName, "DiskCacheHit")

var statsCounters = []*Counter{
	S3ListObjects,
	S3HeadObject,
	S3PutObject,
	S3GetObject,
	S3DeleteObjects,
	S3DeleteObject,

	CacheRead,
	CacheHit,
	MemCacheRead,
	MemCacheHit,
	DiskCacheRead,
	DiskCacheHit,
}

// Gather returns the snapshot of all the statsFamily in the registry
func Gather() (statsFamilies map[string][]zap.Field) {

	for _, familyName := range familyNames {
		statsFamilies[familyName] = []zap.Field{}
	}

	for _, statsCounter := range statsCounters {
		stat := zap.Any(statsCounter.name, statsCounter.Load())
		statsCounter.MergeAndReset()
		statsFamilies[statsCounter.fName] = append(statsFamilies[statsCounter.fName], stat)
	}
	return
}
