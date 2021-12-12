package golanglibs

import (
	"github.com/mmcdole/gofeed"
)

type rssConfig struct {
	proxy        string
	retryOnError bool
	timeout      int
}

func getRSS(url string, config ...rssConfig) *gofeed.Feed {
	fp := gofeed.NewParser()

	var cfg HttpConfig
	var retryOnError bool
	if len(config) != 0 {
		cfg.httpProxy = config[0].proxy
		retryOnError = config[0].retryOnError
		if config[0].timeout != 0 {
			cfg.timeout = config[0].timeout
		}
	}

	var feed *gofeed.Feed
	var err error
	if retryOnError {
		for {
			if err := try(func() {
				content := httpGet(url, cfg).content

				// lg.trace("获取到的内容是:", content)

				feed, err = fp.ParseString(content)
				panicerr(err)
			}).Error; err == nil {
				break
			} else {
				// lg.trace("获取RSS失败:", err)
				sleep(1)
				// lg.trace("重试")
			}
		}
	} else {
		content := httpGet(url, cfg).content

		// lg.trace("获取到的内容是:", content)

		feed, err = fp.ParseString(content)
		panicerr(err)
	}

	return feed
}