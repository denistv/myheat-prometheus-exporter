package services

import (
	"testing"
	"time"
)

func TestTariffSelector_Select(t1 *testing.T) {
	type fields struct {
		timeNowFunc func() time.Time
		tariffs     []Tariff
	}
	tests := []struct {
		name   string
		fields fields
		want   TariffType
	}{
		{
			name: "тариф по-умолчанию",
			fields: fields{
				timeNowFunc: func() time.Time { return time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC) },
				tariffs:     []Tariff{},
			},
			want: TariffOne,
		},
		{
			name: "ночной тариф: начало интервала",
			fields: fields{
				timeNowFunc: func() time.Time { return time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC) },
				tariffs:     []Tariff{NewNightTariff(0, 10)},
			},
			want: TariffTwo,
		},
		{
			name: "ночной тариф: середина интервала",
			fields: fields{
				timeNowFunc: func() time.Time { return time.Date(2024, time.January, 1, 5, 0, 0, 0, time.UTC) },
				tariffs:     []Tariff{NewNightTariff(0, 10)},
			},
			want: TariffTwo,
		},
		{
			name: "ночной тариф: конец интервала",
			fields: fields{
				timeNowFunc: func() time.Time { return time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC) },
				tariffs:     []Tariff{NewNightTariff(0, 10)},
			},
			want: TariffOne,
		},
		{
			name: "ночной тариф: выход за конец интервала",
			fields: fields{
				timeNowFunc: func() time.Time { return time.Date(2024, time.January, 1, 11, 0, 0, 0, time.UTC) },
				tariffs:     []Tariff{NewNightTariff(0, 10)},
			},
			want: TariffOne,
		},
		{
			name: "ночной тариф: вторая граница уходит на следующий день",
			fields: fields{
				timeNowFunc: func() time.Time { return time.Date(2024, time.January, 1, 23, 0, 0, 0, time.UTC) },
				tariffs:     []Tariff{NewNightTariff(22, 7)},
			},
			want: TariffTwo,
		},
	}

	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TariffSelector{
				timeNowFunc: tt.fields.timeNowFunc,
				tariffs:     tt.fields.tariffs,
			}
			if got := t.Select(); got != tt.want {
				t1.Errorf("Select() = %v, want %v", got, tt.want)
			}
		})
	}
}
