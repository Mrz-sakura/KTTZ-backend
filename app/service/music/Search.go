package music

import (
	"app-bff/app/service/_dto/music"
	"app-bff/pkg/errorss"
	"app-bff/pkg/res"
	"github.com/gin-gonic/gin"
)

type Search struct {
	HTTPGETSearch string `path:"/music/search" method:"GET"`
}

// 控制器的方法 分页查询
func (t *Search) GETSearch(c *gin.Context) {
	query, err := t.VerifySearchParams(c)
	if err != nil {
		res.Error(c, err)
		return
	}
	res.Ok(c, query)
}

func (*Search) VerifySearchParams(c *gin.Context) (*music.SearchRequest, error) {
	query := &music.SearchRequest{Keywords: c.Query("keywords")}
	if query.Keywords == "" {
		return nil, errorss.NewWithCodeMsg(errorss.BAD_PARAMETER, "keywords不能为空")
	}
	return query, nil
}
