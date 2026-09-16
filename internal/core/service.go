package core

import (
	"log"
	"sync"

	"github.com/shopspring/decimal"
	"github.com/werastine/CryptoNotifier/pkg/pb"
)

type CryptoService struct {
	Target []TargetInfo
	mtx    sync.RWMutex
}

type TargetInfo struct {
	Price      string
	Symbol     string
	ID         string
	Condition  pb.Condition // <= or >=
	ClientChan chan *pb.PriceTick
}

func NewCryptoService() *CryptoService {
	return &CryptoService{
		Target: make([]TargetInfo, 0),
	}
}

func (cs *CryptoService) AddTarget(
	price string,
	symbol string,
	id string,
	condition pb.Condition,
	clientChan chan *pb.PriceTick,
) error {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	cs.Target = append(cs.Target, TargetInfo{
		Price:      price,
		Symbol:     symbol,
		ID:         id,
		Condition:  condition,
		ClientChan: clientChan,
	})
	return nil
}

func (cs *CryptoService) RemoveTarget(UUID string) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	for i, item := range cs.Target {
		if item.ID == UUID {
			lastIdx := len(cs.Target) - 1

			cs.Target[i] = cs.Target[lastIdx]

			var zero TargetInfo
			cs.Target[lastIdx] = zero

			cs.Target = cs.Target[:lastIdx]
			log.Printf("[INFO] Deleted target: %s from slice", UUID)
			return
		}
	}
	log.Printf("[WARN] Target %s is not found", UUID)
}

func (cs *CryptoService) FindTarget(symbol string, price string) []chan *pb.PriceTick {
	cs.mtx.RLock()
	defer cs.mtx.RUnlock()

	var tickStorage []chan *pb.PriceTick

	tickPrice, err := decimal.NewFromString(price)
	if err != nil {
		log.Printf("[ERROR] invalid input price string: %v", err)
		return nil
	}

	for _, j := range cs.Target {
		if j.Symbol != symbol {
			continue
		}

		targetPrice, err := decimal.NewFromString(j.Price)
		if err != nil {
			continue
		}

		if tickPrice.LessThanOrEqual(targetPrice) && pb.Condition_CONDITION_BELOW == j.Condition {
			tickStorage = append(tickStorage, j.ClientChan)
			continue
		}
		if tickPrice.GreaterThanOrEqual(targetPrice) && pb.Condition_CONDITION_ABOVE == j.Condition {
			tickStorage = append(tickStorage, j.ClientChan)
			continue
		}
	}
	return tickStorage
}
