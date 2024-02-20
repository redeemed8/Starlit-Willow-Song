package models

import (
	"fmt"
	"gorm.io/gorm"
	"jcpd.cn/user/internal/options"
	"strings"
)

var PointInfoDao PointInfoDao_
var PointInfoUtil PointInfoUtil_

type PointInfoDao_ struct{ DB *gorm.DB }
type PointInfoUtil_ struct{}

func NewPointInfoDao() {
	PointInfoDao = PointInfoDao_{DB: options.C.DB}
}

type PointInfo struct {
	Id    uint32 `gorm:"primaryKey"`
	Point Point  `gorm:"type:geometry;not null;index:idx_location"`
}

// Point 定义地理位置的结构体
type Point struct {
	X float64 //	经度
	Y float64 //	纬度
}

const PointTableName = "5918_pointinfo"

// TableName 表名
func (table *PointInfo) TableName() string {
	return PointTableName
}

// CreateTable 创建表
func (info *PointInfoDao_) CreateTable() {
	_ = info.DB.AutoMigrate(&PointInfo{})
}

// CreatePointInfo 创建位置信息
func (info *PointInfoDao_) CreatePointInfo(pointInfo PointInfo) error {
	sqlSlice := []string{
		"insert into",
		pointInfo.TableName(),
		"values",
		fmt.Sprintf("(%d,POINT(%f,%f))", pointInfo.Id, pointInfo.Point.X, pointInfo.Point.Y),
	}
	sql_ := strings.Join(sqlSlice, " ")
	return info.DB.Exec(sql_).Error
}

// CheckPointIsExists 检查某个 id信息是否存在
func (info *PointInfoDao_) CheckPointIsExists(id uint32) (bool, error) {
	var count_ int64
	err := info.DB.Model(&PointInfo{}).Where("id = ?", id).Count(&count_).Error
	return count_ == 1, err
}

// UpdatePosById 根据 id修改位置信息
func (info *PointInfoDao_) UpdatePosById(pointInfo PointInfo) error {
	sqlSlice := []string{
		"update",
		pointInfo.TableName(),
		fmt.Sprintf("set point = POINT(%f,%f)", pointInfo.Point.X, pointInfo.Point.Y),
		fmt.Sprintf("where id = %d", pointInfo.Id),
	}
	sql_ := strings.Join(sqlSlice, " ")
	return info.DB.Exec(sql_).Error
}

//	---------------------------------------------

type nearbyUser_ struct {
	Id       uint32
	Distance float64
}

const KM = 1000
const DistanceMAX = 500 * KM //  最大范围 - 500km

type IdDisMap map[uint32]string

func (map_ *IdDisMap) Keys() []uint32 {
	var keys []uint32
	for k := range *map_ {
		keys = append(keys, k)
	}
	return keys
}

func (info *PointInfoDao_) GetUserByDistance(origin Point, distance int, limit int, offset int) (IdDisMap, error) {
	if distance > DistanceMAX {
		distance = DistanceMAX
	}
	sqlSlice := []string{
		"select id,distance from",
		fmt.Sprintf("(select *,ST_Distance_Sphere(POINT(%f,%f),point) as distance", origin.X, origin.Y),
		fmt.Sprintf("from %s) as distances", PointTableName),
		fmt.Sprintf("where distance <= %d order by distance limit %d offset %d", distance, limit, offset),
	}
	sql_ := strings.Join(sqlSlice, " ")
	var nearbyUser []nearbyUser_
	err := info.DB.Raw(sql_).Scan(&nearbyUser).Error

	map_ := make(IdDisMap)
	for _, near := range nearbyUser {
		var dis string
		if near.Distance < 1000.000000 {
			dis = fmt.Sprintf("%dm", int(near.Distance))
		} else {
			dis = fmt.Sprintf("%dkm", int(near.Distance)/1000)
		}
		map_[near.Id] = dis
	}
	return map_, err
}

//	---------------------------------------------

func (util *PointInfoUtil_) CheckPointXY(x float64, y float64) bool {
	return x > -180 && x <= 180 && y >= -90 && y <= 90 && x != 0 || y != 0
}

func (util *PointInfoUtil_) MakePoint(x float64, y float64) Point {
	return Point{x, y}
}
