package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"golang.org/x/exp/rand"
	"strconv"
	"strings"
	"time"
)

func MakeCodeWithNumber(length int, symbol int) (ret string) {
	if length <= 0 {
		return ""
	}
	if length > 25 {
		length = 25
	}
	rng := rand.New(rand.NewSource(uint64(time.Now().UnixNano() + int64(symbol))))
	for i := 0; i < length; i++ {
		num := rng.Intn(10)
		ret += strconv.Itoa(num)
	}
	return ret
}

func Md5Sum(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// ParseListToUint 将id串转换为uint32的集合
func ParseListToUint(list string) []uint32 {
	idStrArr := strings.Split(list, ",")
	if list == "" || len(idStrArr) == 0 {
		return make([]uint32, 0)
	}
	ids := make([]uint32, 0) //	结果
	for _, idStr := range idStrArr {
		//	转换为 uint32
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			continue
		}
		ids = append(ids, uint32(id))
	}
	return ids
}

// RemoveIdFromList 删除切片中指定位置的元素
func RemoveIdFromList(idList *[]uint32, index int) {
	if index >= len(*idList) || index < 0 {
		return
	}
	*idList = append((*idList)[:index], (*idList)[index+1:]...)
}

// JoinUint32 将 uint32数组转化为字符串
func JoinUint32(ids []uint32) string {
	var idsStr string
	for i, id := range ids {
		idsStr = fmt.Sprintf("%s%d", idsStr, id)
		if i+1 == len(ids) {
			break
		}
		idsStr += ","
	}
	return idsStr
}
