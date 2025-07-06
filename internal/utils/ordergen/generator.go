package ordergen

import (
	"github.com/divanov-web/gophermart/internal/model"
	"math/rand"
	"time"
)

var (
	types  = []string{"чайник", "плита", "кофемашина", "холодильник"}
	brands = []string{"Bork", "Bosh", "LG"}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomGoods случайная генерация состава заказа
func GenerateRandomGoods() []model.OrderGoods {
	count := rand.Intn(4) + 1 // от 1 до 4 товаров
	goods := make([]model.OrderGoods, 0, count)

	for i := 0; i < count; i++ {
		description := types[rand.Intn(len(types))] + " " + brands[rand.Intn(len(brands))]
		price := rand.Intn(49001) + 1000 // от 1000 до 50000
		goods = append(goods, model.OrderGoods{
			Description: description,
			Price:       price,
		})
	}

	return goods
}
