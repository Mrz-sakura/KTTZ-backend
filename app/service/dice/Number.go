package dice

import (
	"app-bff/app/service/_dto/dice"
	"app-bff/mod"
	"app-bff/pkg/config"
	"app-bff/pkg/res"
	"app-bff/pkg/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
)

type Number struct {
	HTTPPOSTNumber string `path:"/dice/number" method:"POST"`
}

// 控制器的方法 分页查询
func (t *Number) POSTNumber(c *gin.Context) {
	query, err := t.VerifySearchParams(c)
	if err != nil {
		res.Error(c, err)
		return
	}

	diceState, err := t.GeneratorValue(c, query.LockedIndexes)
	if err != nil {
		res.Error(c, err)
		return
	}
	fmt.Println(diceState.DiceValues)
	res.Ok(c, diceState)
}

func (*Number) VerifySearchParams(c *gin.Context) (*dice.NumberRequest, error) {
	query := &dice.NumberRequest{}

	if err := c.BindJSON(query); err != nil {
		return nil, err
	}

	return query, nil
}

func (t *Number) GeneratorValue(c *gin.Context, lockedIndexes []string) (*dice.DiceState, error) {
	rc, err := mod.GetRedisClient()
	if err != nil {
		return nil, err
	}

	// TODO 1是写死的,后续换成userid
	key := fmt.Sprintf("%s%d", config.GetString("redis_key.dice_key"), 1)

	diceState := &dice.DiceState{DiceValues: make([]int, 5)}

	// 获取redis的值,如果没有,代表是新的一轮
	if val, err := rc.Get(c, key).Result(); err == nil {
		err = json.Unmarshal([]byte(val), &diceState.DiceValues)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < 5; i++ {
		// 如果当前索引在锁定索引列表中，则不生成新的随机数
		if utils.ArrayStrContainsInt(lockedIndexes, i) {
			continue
		}
		diceState.DiceValues[i] = rand.Intn(6) + 1
	}

	setv, err := json.Marshal(diceState.DiceValues)
	if err != nil {
		return nil, err
	}

	_, err = rc.Set(c, key, setv, 0).Result()
	if err != nil {
		return nil, err
	}

	return diceState, nil
}
