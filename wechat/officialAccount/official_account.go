package officialAccount

import (
	"github.com/ArtisanCloud/PowerLibs/v3/cache"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"
	"github.com/zjutjh/mygo/nesty"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/wechat"
)

// New 以指定配置创建实例
// https://powerwechat.artisan-cloud.com/zh/official-account/
func New(conf Config) (*officialAccount.OfficialAccount, error) {
	l := nlog.Pick(conf.Log)
	client := nesty.Pick(conf.Resty)

	var kernelCache cache.CacheInterface
	switch conf.Cache.Driver {
	case CacheDriverRedis:
		rdb := nedis.Pick(conf.Redis)
		gr := cache.NewGRedis(&redis.UniversalOptions{})
		gr.Pool = rdb
		kernelCache = gr
	case CacheDriverMemory:
		kernelCache = cache.NewMemCache(conf.Cache.MemCache.Namespace, conf.Cache.MemCache.DefaultLifeTime, conf.Cache.MemCache.Prefix)
	default:
		kernelCache = nil
	}

	uc := &officialAccount.UserConfig{
		AppID:             conf.AppID,
		Secret:            conf.Secret,
		Token:             conf.Token,
		AESKey:            conf.AESKey,
		StableTokenMode:   conf.StableTokenMode,
		ForceRefresh:      conf.ForceRefresh,
		RefreshToken:      conf.RefreshToken,
		ComponentAppID:    conf.ComponentAppID,
		ComponentAppToken: conf.ComponentAppToken,
		ResponseType:      conf.ResponseType,
		Http: officialAccount.Http{
			Timeout:   conf.Http.Timeout,
			BaseURI:   conf.Http.BaseURI,
			ProxyURI:  conf.Http.ProxyURI,
			Transport: client.GetClient().Transport,
		},
		Log: officialAccount.Log{
			Driver: wechat.NewLogger(l),
		},
		Cache:     kernelCache,
		HttpDebug: conf.HttpDebug,
		Debug:     conf.Debug,
		OAuth: officialAccount.OAuth{
			Callback: conf.OAuth.Callback,
			Scopes:   conf.OAuth.Scopes,
		},
	}

	oa, err := officialAccount.NewOfficialAccount(uc)
	if err != nil {
		return nil, err
	}

	return oa, nil
}
