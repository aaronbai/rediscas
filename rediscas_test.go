package rediscas

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/gomodule/redigo/redis"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	c *Conn
)

const (
	testKey = "testCasKey"
	testVal = "testCasVal"
)

func TestNotExist(t *testing.T) {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(fmt.Sprintf("dial err:%+v", err))
	}
	c = &Conn{conn}
	defer c.Close()

	Convey("TestNotExist", t, func() {
		val, cas, err := c.Get(testKey)
		So(err, ShouldEqual, ErrNotExist)
		So(cas, ShouldEqual, 0)
		So(val, ShouldEqual, "")
	})
}
func TestSetNil(t *testing.T) {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(fmt.Sprintf("dial err:%+v", err))
	}
	c = &Conn{conn}
	defer c.Close()

	Convey("TestSetNil", t, func() {
		err := c.Set(testKey, "", -1)
		So(err, ShouldBeNil)
		val, cas, err := c.Get(testKey)
		So(err, ShouldBeNil)
		So(cas, ShouldEqual, 1)
		So(val, ShouldEqual, "")
		// 清理
		err = c.Del(testKey)
		So(err, ShouldBeNil)
	})
}
func TestCasSet(t *testing.T) {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(fmt.Sprintf("dial err:%+v", err))
	}
	c = &Conn{conn}
	defer c.Close()

	Convey("TestCasSet", t, func() {
		Convey("TestSet", func() {
			for i := 0; i < 10; i++ {
				err := c.Set(testKey, testVal, i)
				So(err, ShouldBeNil)
				val, cas, err := c.Get(testKey)
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testVal)
				So(cas, ShouldEqual, i+1)
			}
			// 清理环境
			err := c.Del(testKey)
			So(err, ShouldBeNil)
		})
		Convey("TestGetSet", func() {
			// 初始化环境
			err := c.Set(testKey, testVal, 10)
			So(err, ShouldBeNil)
			_, cas, err := c.Get(testKey)
			So(err, ShouldBeNil)
			Convey("testForceGetSet", func() {
				err = c.Set(testKey, "testVal1", -1)
				So(err, ShouldBeNil)
			})
			Convey("testGetSet", func() {
				err = c.Set(testKey, "testVal1", cas)
				So(err, ShouldBeNil)
			})
			val, casNew, err := c.Get(testKey)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, "testVal1")
			So(casNew, ShouldEqual, cas+1)
			// 清理测试环境
			err = c.Del(testKey)
			So(err, ShouldBeNil)
		})
	})
}
func TestBatchGet(t *testing.T) {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(fmt.Sprintf("dial err:%+v", err))
	}
	c = &Conn{conn}
	defer c.Close()

	Convey("TestBatchGet", t, func() {
		Convey("TestBatchGet", func() {
			keys := make([]string, 0)
			for i := 0; i < 10; i++ {
				key := testKey + strconv.Itoa(i)
				keys = append(keys, key)
				err := c.Set(key, testVal, -1)
				So(err, ShouldBeNil)
			}
			vals, cas, err := c.BatchGet(keys)
			So(err, ShouldBeNil)
			val, ok := vals[keys[0]]
			So(ok, ShouldBeTrue)
			So(val, ShouldEqual, testVal)
			So(cas[keys[0]], ShouldEqual, 1)
			// 清理环境
			for _, key := range keys {
				err := c.Del(key)
				So(err, ShouldBeNil)
			}
		})
	})
}
