package db

import (
	"reflect"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func TestDB(t *testing.T) {
	tests := []struct {
		name string
		want *gorm.DB
	}{
		{
			name: "Test",
			want: db,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DB(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB() = %v, want %v", got, tt.want)
			}
		})
	}
}
