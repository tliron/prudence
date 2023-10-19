package badger

import (
	"fmt"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	badgeroptions "github.com/dgraph-io/badger/v4/options"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/platform"
)

func NewDBOptions(path string, config any) (badger.Options, error) {
	options := badger.DefaultOptions(path)

	config_ := ard.With(config).ConvertSimilar().NilMeansZero()

	if config__, ok := config_.StringMap(); ok {
		platform.ValidateConfigKeys("Badger DB", config__,
			"syncWrites",
			"numVersionsToKeep",
			"readOnly",
			"compression",
			"inMemory",
			"metricsEnabled",
			"numGoroutines",
			"memTableSize",
			"baseTableSize",
			"baseLevelSize",
			"levelSizeMultiplier",
			"tableSizeMultiplier",
			"maxLevels",
			"vLogPercentile",
			"valueThreshold",
			"numMemtables",
			"blockSize",
			"bloomFalsePositive",
			"blockCacheSize",
			"indexCacheSize",
			"numLevelZeroTables",
			"numLevelZeroTablesStall",
			"valueLogFileSize",
			"valueLogMaxEntries",
			"numCompactors",
			"compactL0OnClose",
			"lmaxCompaction",
			"zstdCompressionLevel",
			"verifyValueChecksum",
			"encryptionKey",
			"encryptionKeyRotationDuration",
			"bypassLockGuard",
			"checksumVerificationMode",
			"detectConflicts",
			"namespaceOffset",
			"externalMagicVersion",
		)
	} else {
		return options, fmt.Errorf("not a config: %T", config_.Value)
	}

	if syncWrites, ok := config_.Get("syncWrites").Boolean(); ok {
		options.SyncWrites = syncWrites
	}

	if numVersionsToKeep, ok := config_.Get("numVersionsToKeep").Integer(); ok {
		options.NumVersionsToKeep = int(numVersionsToKeep)
	}

	if readOnly, ok := config_.Get("readOnly").Boolean(); ok {
		options.ReadOnly = readOnly
	}

	if compression, ok := config_.Get("compression").String(); ok {
		switch compression {
		case "none":
			options.Compression = badgeroptions.None
		case "snappy":
			options.Compression = badgeroptions.Snappy
		case "zstd":
			options.Compression = badgeroptions.ZSTD
		default:
			return options, fmt.Errorf("unsupported compression: %s", compression)
		}
	}

	if inMemory, ok := config_.Get("inMemory").Boolean(); ok {
		options.InMemory = inMemory
	}

	if metricsEnabled, ok := config_.Get("metricsEnabled").Boolean(); ok {
		options.MetricsEnabled = metricsEnabled
	}

	if numGoroutines, ok := config_.Get("numGoroutines").Integer(); ok {
		options.NumGoroutines = int(numGoroutines)
	}

	// Fine tuning options

	if memTableSize, ok := config_.Get("memTableSize").Integer(); ok {
		options.MemTableSize = memTableSize
	}

	if baseTableSize, ok := config_.Get("baseTableSize").Integer(); ok {
		options.BaseTableSize = baseTableSize
	}

	if baseLevelSize, ok := config_.Get("baseLevelSize").Integer(); ok {
		options.BaseLevelSize = baseLevelSize
	}

	if levelSizeMultiplier, ok := config_.Get("levelSizeMultiplier").Integer(); ok {
		options.LevelSizeMultiplier = int(levelSizeMultiplier)
	}

	if tableSizeMultiplier, ok := config_.Get("tableSizeMultiplier").Integer(); ok {
		options.TableSizeMultiplier = int(tableSizeMultiplier)
	}

	if maxLevels, ok := config_.Get("maxLevels").Integer(); ok {
		options.MaxLevels = int(maxLevels)
	}

	if vLogPercentile, ok := config_.Get("vLogPercentile").Float(); ok {
		options.VLogPercentile = vLogPercentile
	}

	if valueThreshold, ok := config_.Get("valueThreshold").Integer(); ok {
		options.ValueThreshold = valueThreshold
	}

	if numMemtables, ok := config_.Get("numMemtables").Integer(); ok {
		options.NumMemtables = int(numMemtables)
	}

	if blockSize, ok := config_.Get("blockSize").Integer(); ok {
		options.BlockSize = int(blockSize)
	}

	if bloomFalsePositive, ok := config_.Get("bloomFalsePositive").Float(); ok {
		options.BloomFalsePositive = bloomFalsePositive
	}

	if blockCacheSize, ok := config_.Get("blockCacheSize").Integer(); ok {
		options.BlockCacheSize = blockCacheSize
	}

	if indexCacheSize, ok := config_.Get("indexCacheSize").Integer(); ok {
		options.IndexCacheSize = indexCacheSize
	}

	if numLevelZeroTables, ok := config_.Get("numLevelZeroTables").Integer(); ok {
		options.NumLevelZeroTables = int(numLevelZeroTables)
	}

	if numLevelZeroTablesStall, ok := config_.Get("numLevelZeroTablesStall").Integer(); ok {
		options.NumLevelZeroTablesStall = int(numLevelZeroTablesStall)
	}

	if valueLogFileSize, ok := config_.Get("valueLogFileSize").Integer(); ok {
		options.ValueLogFileSize = valueLogFileSize
	}

	if valueLogMaxEntries, ok := config_.Get("valueLogMaxEntries").UnsignedInteger(); ok {
		options.ValueLogMaxEntries = uint32(valueLogMaxEntries)
	}

	if numCompactors, ok := config_.Get("numCompactors").Integer(); ok {
		options.NumCompactors = int(numCompactors)
	}

	if compactL0OnClose, ok := config_.Get("compactL0OnClose").Boolean(); ok {
		options.CompactL0OnClose = compactL0OnClose
	}

	if lmaxCompaction, ok := config_.Get("lmaxCompaction").Boolean(); ok {
		options.LmaxCompaction = lmaxCompaction
	}

	if zstdCompressionLevel, ok := config_.Get("zstdCompressionLevel").Integer(); ok {
		options.ZSTDCompressionLevel = int(zstdCompressionLevel)
	}

	if verifyValueChecksum, ok := config_.Get("verifyValueChecksum").Boolean(); ok {
		options.VerifyValueChecksum = verifyValueChecksum
	}

	// Encryption related options

	if encryptionKey, ok := config_.Get("encryptionKey").Bytes(); ok {
		options.EncryptionKey = encryptionKey
	}

	if encryptionKeyRotationDuration, ok := config_.Get("encryptionKeyRotationDuration").Float(); ok {
		options.EncryptionKeyRotationDuration = time.Duration(encryptionKeyRotationDuration * float64(time.Second))
	}

	if bypassLockGuard, ok := config_.Get("bypassLockGuard").Boolean(); ok {
		options.BypassLockGuard = bypassLockGuard
	}

	if checksumVerificationMode, ok := config_.Get("checksumVerificationMode").String(); ok {
		switch checksumVerificationMode {
		case "noVerification":
			options.ChecksumVerificationMode = badgeroptions.NoVerification
		case "onTableRead":
			options.ChecksumVerificationMode = badgeroptions.OnTableRead
		case "onBlockRead":
			options.ChecksumVerificationMode = badgeroptions.OnBlockRead
		case "onTableAndBlockRead":
			options.ChecksumVerificationMode = badgeroptions.OnTableAndBlockRead
		default:
			return options, fmt.Errorf("unsupported checksumVerificationMode: %s", checksumVerificationMode)
		}
	}

	if detectConflicts, ok := config_.Get("detectConflicts").Boolean(); ok {
		options.DetectConflicts = detectConflicts
	}

	if namespaceOffset, ok := config_.Get("namespaceOffset").Integer(); ok {
		options.NamespaceOffset = int(namespaceOffset)
	}

	if externalMagicVersion, ok := config_.Get("externalMagicVersion").UnsignedInteger(); ok {
		options.ExternalMagicVersion = uint16(externalMagicVersion)
	}

	options.Logger = Logger{}
	return options, nil
}

func NewIteratorOptions(config any) (badger.IteratorOptions, error) {
	options := badger.DefaultIteratorOptions

	config_ := ard.With(config).ConvertSimilar().NilMeansZero()

	if config__, ok := config_.StringMap(); ok {
		platform.ValidateConfigKeys("Badger Iterator", config__,
			"prefix",
			"prefetchValues",
			"prefetchSize",
			"reverse",
			"allVersions",
			"internalAccess",
		)
	} else {
		return options, fmt.Errorf("not a config: %T", config_.Value)
	}

	if prefix, ok := config_.Get("prefix").String(); ok {
		options.Prefix = util.ToBytes(prefix)
	}

	if prefetchValues, ok := config_.Get("prefetchValues").Boolean(); ok {
		options.PrefetchValues = prefetchValues
	}

	if prefetchSize, ok := config_.Get("prefetchSize").Integer(); ok {
		options.PrefetchSize = int(prefetchSize)
	}

	if reverse, ok := config_.Get("reverse").Boolean(); ok {
		options.Reverse = reverse
	}

	if allVersions, ok := config_.Get("allVersions").Boolean(); ok {
		options.AllVersions = allVersions
	}

	if internalAccess, ok := config_.Get("internalAccess").Boolean(); ok {
		options.InternalAccess = internalAccess
	}

	return options, nil
}
