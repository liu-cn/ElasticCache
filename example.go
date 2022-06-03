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
