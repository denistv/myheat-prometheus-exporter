package services

import (
	"strconv"
	"time"
)

type TariffType int

func (t TariffType) String() string {
	return strconv.FormatInt(int64(t), 10)
}

const TariffOne TariffType = 1 // default tariffType
const TariffTwo TariffType = 2 // night

func NewTariff(tariff TariffType, from, to int) Tariff {
	return Tariff{
		tariffType: tariff,
		from:       from,
		to:         to,
	}
}

type Tariff struct {
	tariffType TariffType
	from       int
	to         int
}

func NewNightTariff(from, to int) Tariff {
	return Tariff{
		tariffType: TariffTwo,
		from:       from,
		to:         to,
	}
}

func NewDefaultTariff() Tariff {
	return Tariff{
		tariffType: TariffOne,
		from:       0, // с 00:00 текущего дня
		to:         0, // до 00:00 следующего дня
	}
}

func NewTariffSelector(timeNowFunc func() time.Time, tariffs []Tariff) *TariffSelector {
	return &TariffSelector{
		timeNowFunc: timeNowFunc,
		tariffs:     tariffs,
	}
}

type TariffSelector struct {
	timeNowFunc func() time.Time
	tariffs     []Tariff
}

// Select возвращает первый подходящий тариф. Если ни один из тарифов не выбрался, возвращается дефолтный
func (t *TariffSelector) Select() TariffType {
	now := t.timeNowFunc()

	for _, tariff := range t.tariffs {
		intervalMatched := false
		oneDay := (tariff.from - tariff.to) < 0

		if oneDay {
			// итнтервал в рамках текущего дня
			// пример: from=10; to=22
			fromMatched := now.Hour() >= tariff.from
			toMatched := now.Hour() < tariff.to
			intervalMatched = fromMatched && toMatched
		} else {
			// интервал может быть в рамках двух соседних дней
			// Пример: from=22; now=23; to=7

			// from
			if (tariff.from <= now.Hour()) && (now.Hour() <= 24) {
				intervalMatched = true
			}
			// to
			if (0 <= now.Hour()) && (now.Hour() < tariff.to) {
				intervalMatched = true
			}
		}

		if intervalMatched {
			return tariff.tariffType
		}
	}

	return TariffOne
}
