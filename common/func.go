package common

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

const TimeFormat = "2006-01-02 15:04:05"

var location *time.Location

func init() {
	location = time.FixedZone("CST", 8*3600)
	rand.Seed(time.Now().UnixNano())
}

//GetRootPath 获取项目根目录 (即 main.go的所在位置)
func GetRootPath(relPath ...string) string {
	pwd, _ := os.Getwd()
	if len(relPath) > 0 {
		pwd = filepath.Join(pwd, relPath[0])
	}
	return pwd
}

func UcFirst(str string) string {
	if len(str) < 1 {
		return ""
	}
	strArr := []rune(str)
	if strArr[0] >= 97 && strArr[0] <= 122 {
		strArr[0] -= 32
	}
	return string(strArr)
}

func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				index = i
				exists = true
				return
			}
		}
	}
	return
}

//GetMd5 md5加密字符串
func GetMd5(str string) string {
	md := md5.New()
	md.Write([]byte(str))
	return hex.EncodeToString(md.Sum(nil))
}
