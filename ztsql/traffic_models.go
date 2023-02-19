package ztsql

import "gitee.com/sienectagv/gozk/zsql"

// type Model = gorm.Model

type NoparkingSpots struct {
	SpotCode string `gorm:"type:varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci;primarykey"`
}

type AllHoldSpots struct {
	Model zsql.Model `gorm:"embedded"`
}

type ResourceForSpots struct {
	Model           zsql.Model `gorm:"embedded"`
	Code            string     `gorm:"type:varchar(64); unique; not null;"`
	Type            int32      `gorm:"not null;default:0;"`
	MaxHoldNumber   int32      `gorm:"not null;default:0;"`
	Extend          string     `gorm:"type:text;"`
	Name            string     `gorm:"type:varchar(255);not null; default:'';"`
	PointsSet       string     `gorm:"type:text;"`                       //join with ',' when multiple codes are specified
	RequestByRobots string     `gorm:"type:text;"`                       //
	HoldByRobots    string     `gorm:"type:text;"`                       //join with ',' when multiple robots are specified
	HoldenByUs      int32      `gorm:"type:int(11);not null;default:1;"` //0:未占用，1:被我们占用
	MapCode         string     `gorm:"type:varchar(255);"`               //
	Floors          string     `gorm:"type:text;"`                       //电梯资源的楼层资料格式是分号分隔不同楼层，冒号分隔楼层数字和点位名称，同一个楼层有多个电梯用逗号分隔\n1:spot_f1_a,spot_f1_b;2:spot_f2_a,spot_f2_b
	Address         string     `gorm:"type:varchar(64);"`
	ClientID        int32      `gorm:"type:int(11);"` //负责控股之的客户端ID
}

type ActionForSpots struct {
	Model zsql.Model `gorm:"embedded"`
	// Action int32      `gorm:"type:int(11);not null;"`//Use ID
	Name       string `gorm:"type:varchar(128);"`
	ResourceId ResourceForSpots
	StepByStep int32             `gorm:"type:int(11);not null;default:0;"`
	Triggers   []TriggerForSpots `gorm:"many2many:action_triggers;"` //many2many
}

type TriggerForSpots struct {
	Model     zsql.Model `gorm:"embedded"`
	StartSpot string     `gorm:"type:varchar(64);not null;"`
	EndSpot   string     `gorm:"type:varchar(64);not null;"`
}
