# 带Cas操作的Redis

## Cas 简介
cas俗称乐观锁，是解决并发冲突的一种方案。
详细介绍:(TODO)

## 具体实现
使用redis本身lua脚本特性实现，参见代码。

## 使用方法
提供如下几个接口，详细示例参考rediscas_test.go:
```

// Get操作, 入参: key 出参: val, cas, err
func (r *CasConn) Get(key string) (string, int, error)

// GetProto, 同Get, 自动解析pb结构体
func (r *CasConn) GetProto(key string, msg proto.Message) (int, error)

// Set操作，入参: key, val, cas, 出参: err
func (r *CasConn) Set(key, val string, cas int) error

// SetProto, 同Set, 自动unmarshal pb结构体
func (r *CasConn) SetProto(key string, info proto.Message, cas int)

// Del操作，入参: key, 出参: err
func (r *CasConn) Del(key string) error

// BatchGet 批查，入参: keys, 出参: msgs、cas、err
// (如果查询的key在redis中不存在，则msgs和cas中没有对应key)
func BatchGet(keys []string, msgs map[string]string, cas map[string]uint64) error

// BatchGetProto 批查
func BatchGetProto(keys []string, msgs map[string]proto.Message, cas map[string]uint64) error
```
