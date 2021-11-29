package memory

import (
	"time"

	"github.com/tliron/prudence/platform"
)

type GetExpirationFunc func(key platform.CacheKey) (time.Time, bool)

//
// CacheGroup
//

type CacheGroup struct {
	Keys       []platform.CacheKey
	Expiration time.Time
}

func (self *CacheGroup) Expired() bool {
	return (len(self.Keys) == 0) || (time.Now().After(self.Expiration))
}

func (self *CacheGroup) Prune(getExpiration GetExpirationFunc) {
	keys := self.Keys
	self.Keys = nil
	self.Expiration = time.Time{}
	for _, key := range keys {
		if expiration, ok := getExpiration(key); ok {
			self.Keys = append(self.Keys, key)
			if expiration.After(self.Expiration) {
				self.Expiration = expiration
			}
		}
	}
}

//
// CacheGroups
//

type CacheGroups map[platform.CacheKey]*CacheGroup

func (self CacheGroups) Add(key platform.CacheKey, cached *platform.CachedRepresentation, getExpiration GetExpirationFunc) {
	for _, name := range cached.Groups {
		var group *CacheGroup
		var ok bool
		if group, ok = self[name]; !ok {
			group = new(CacheGroup)
			self[name] = group
		}
		group.Keys = append(group.Keys, key)
		group.Prune(getExpiration)
	}
}

func (self CacheGroups) Delete(name platform.CacheKey, del func(key platform.CacheKey)) {
	if group, ok := self[name]; ok {
		for _, key := range group.Keys {
			del(key)
		}
	}

	delete(self, name)
}

func (self CacheGroups) Prune(getExpiration GetExpirationFunc) {
	for name, group := range self {
		group.Prune(getExpiration)
		if group.Expired() {
			log.Debugf("pruning group: %s", name)
			delete(self, name)
		}
	}
}
