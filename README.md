# ElasticCache
golang本地时效缓存。

## 简介

有时候我们遇到一些高并发读的业务时候，而这些数据又是变化不大，或者对变化的实时性要求不是很高的场景下，我们想尽最大的可能去提高并发量，减少对数据库的读取，

举例一个场景：微博的热搜，属于高热点数据，读取量很大，但是微博的内容基本上是静态的不会变的，除了阅读量，点赞量，收藏，评论，转发，等等，这些统计数据，这些数据都是次要的，没必要每次请求都查询数据库返回最新数据，可以固定时间刷新数据，返回数据统一用缓存，如果每个请求都请求数据库，那么1秒内1万个用户的1万个请求过来了，需要查询1万次数据库，数据库大概率要炸掉，如果不炸掉也基本占用了大部分性能资源了，ElasticCache的目的是，希望帮你将第一次数据库查询到的数据缓存起来，1秒内剩下的9999个人的请求打过来了，不会去查询数据库，会去共享到第一个人的请求数据，这样同样api下1万个请求只需要查询1次数据库就可以了。可以自定义数据存在的有效期，如设置1秒，那么超过1秒就会清理缓存下次再获取时重新创建缓存这时数据已经是最新的数据了，如果没有过期那就直接用缓存，现实场景下1秒的变化并不大，所以没必要把缓存的时间设置的太小。



## 性能对比

这是一个获取评论的接口，数据存储在mongodb，同一个接口，一个用缓存一个没用缓存，性能对比。

#### 精确查询，每次都查数据库

![](https://img-blog.csdnimg.cn/30092cfb1ff1441d9b1482566b0dac67.png#pic_center)

**qps 807，可以看到，每秒撑死了800多个请求（测试多次均是这个数据范围）。**



#### ElasticCache，本地缓存。

![](https://img-blog.csdnimg.cn/28391b6561304fc5bcf527c384a49dd2.png#pic_center)

**可以看到，QPS还差28就140000了，相差100多倍，性能根本不在一个量级。**



## 使用方法：

```go
package ElasticCache

import "time"

var service articleService

func init() {
	service = articleService{
		cache: New(time.Second),
	}
}

type articleService struct {
	cache ElasticCache
}

type articles struct {
	content string
	hot     int //热度
}

func findArticleById(articleID string) (articles, error) {
	//这里只写一个伪代码
	//select * from article while id =1
	//向数据库查询最新数据
	return articles{
		hot:     100,
		content: "文章数据",
	}, nil
}

func GetArticle(articleID string) interface{} {

	data := service.cache.GetAndSet(articleID, time.Second, func(key string) (data interface{}, whetherCache bool) {
		article, err := findArticleById(articleID)
		needCache := true
		if err != nil {
			//whetherCache 代表是否需要缓存该数据，如果true表示此次操作会后的数据会被缓存，如果false则不会缓存该数据
			//data 代表此次查询到的数据，需要被缓存的数据
			return nil, false
		}
		if article.hot < 100 { //这里是自己的业务判断，比如这个hot小于100表示这是一个非常冷门的文章数据，没必要缓存起来，所以needCache = false不会被缓存
			needCache = false
		}
		return article, needCache
	})
	return data
}

```



